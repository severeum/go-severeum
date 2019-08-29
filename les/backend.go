// Copyright 2016 The go-severeum Authors
// This file is part of the go-severeum library.
//
// The go-severeum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-severeum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-severeum library. If not, see <http://www.gnu.org/licenses/>.

// Package les implements the Light Severeum Subprotocol.
package les

import (
	"fmt"
	"sync"
	"time"

	"github.com/severeum/go-severeum/accounts"
	"github.com/severeum/go-severeum/common"
	"github.com/severeum/go-severeum/common/hexutil"
	"github.com/severeum/go-severeum/consensus"
	"github.com/severeum/go-severeum/core"
	"github.com/severeum/go-severeum/core/bloombits"
	"github.com/severeum/go-severeum/core/rawdb"
	"github.com/severeum/go-severeum/core/types"
	"github.com/severeum/go-severeum/sev"
	"github.com/severeum/go-severeum/sev/downloader"
	"github.com/severeum/go-severeum/sev/filters"
	"github.com/severeum/go-severeum/sev/gasprice"
	"github.com/severeum/go-severeum/event"
	"github.com/severeum/go-severeum/internal/sevapi"
	"github.com/severeum/go-severeum/light"
	"github.com/severeum/go-severeum/log"
	"github.com/severeum/go-severeum/node"
	"github.com/severeum/go-severeum/p2p"
	"github.com/severeum/go-severeum/p2p/discv5"
	"github.com/severeum/go-severeum/params"
	rpc "github.com/severeum/go-severeum/rpc"
)

type LightSevereum struct {
	lesCommons

	odr         *LesOdr
	relay       *LesTxRelay
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool

	// Handlers
	peers      *peerSet
	txPool     *light.TxPool
	blockchain *light.LightChain
	serverPool *serverPool
	reqDist    *requestDistributor
	retriever  *retrieveManager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer

	ApiBackend *LesApiBackend

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *sevapi.PublicNetAPI

	wg sync.WaitGroup
}

func New(ctx *node.ServiceContext, config *sev.Config) (*LightSevereum, error) {
	chainDb, err := sev.CreateDB(ctx, config, "lightchaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlockWithOverride(chainDb, config.Genesis, config.ConstantinopleOverride)
	if _, isCompat := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !isCompat {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newPeerSet()
	quitSync := make(chan struct{})

	lsev := &LightSevereum{
		lesCommons: lesCommons{
			chainDb: chainDb,
			config:  config,
			iConfig: light.DefaultClientIndexerConfig,
		},
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		peers:          peers,
		reqDist:        newRequestDistributor(peers, quitSync),
		accountManager: ctx.AccountManager,
		engine:         sev.CreateConsensusEngine(ctx, chainConfig, &config.Sevash, nil, false, chainDb),
		shutdownChan:   make(chan bool),
		networkId:      config.NetworkId,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   sev.NewBloomIndexer(chainDb, params.BloomBitsBlocksClient, params.HelperTrieConfirmations),
	}

	lsev.relay = NewLesTxRelay(peers, lsev.reqDist)
	lsev.serverPool = newServerPool(chainDb, quitSync, &lsev.wg)
	lsev.retriever = newRetrieveManager(peers, lsev.reqDist, lsev.serverPool)

	lsev.odr = NewLesOdr(chainDb, light.DefaultClientIndexerConfig, lsev.retriever)
	lsev.chtIndexer = light.NewChtIndexer(chainDb, lsev.odr, params.CHTFrequencyClient, params.HelperTrieConfirmations)
	lsev.bloomTrieIndexer = light.NewBloomTrieIndexer(chainDb, lsev.odr, params.BloomBitsBlocksClient, params.BloomTrieFrequency)
	lsev.odr.SetIndexers(lsev.chtIndexer, lsev.bloomTrieIndexer, lsev.bloomIndexer)

	// Note: NewLightChain adds the trusted checkpoint so it needs an ODR with
	// indexers already set but not started yet
	if lsev.blockchain, err = light.NewLightChain(lsev.odr, lsev.chainConfig, lsev.engine); err != nil {
		return nil, err
	}
	// Note: AddChildIndexer starts the update process for the child
	lsev.bloomIndexer.AddChildIndexer(lsev.bloomTrieIndexer)
	lsev.chtIndexer.Start(lsev.blockchain)
	lsev.bloomIndexer.Start(lsev.blockchain)

	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		lsev.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	lsev.txPool = light.NewTxPool(lsev.chainConfig, lsev.blockchain, lsev.relay)
	if lsev.protocolManager, err = NewProtocolManager(lsev.chainConfig, light.DefaultClientIndexerConfig, true, config.NetworkId, lsev.eventMux, lsev.engine, lsev.peers, lsev.blockchain, nil, chainDb, lsev.odr, lsev.relay, lsev.serverPool, quitSync, &lsev.wg); err != nil {
		return nil, err
	}
	lsev.ApiBackend = &LesApiBackend{lsev, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.MinerGasPrice
	}
	lsev.ApiBackend.gpo = gasprice.NewOracle(lsev.ApiBackend, gpoParams)
	return lsev, nil
}

func lesTopic(genesisHash common.Hash, protocolVersion uint) discv5.Topic {
	var name string
	switch protocolVersion {
	case lpv1:
		name = "LES"
	case lpv2:
		name = "LES2"
	default:
		panic(nil)
	}
	return discv5.Topic(name + "@" + common.Bytes2Hex(genesisHash.Bytes()[0:8]))
}

type LightDummyAPI struct{}

// Severbase is the address that mining rewards will be send to
func (s *LightDummyAPI) Severbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Coinbase is the address that mining rewards will be send to (alias for Severbase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the severeum package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *LightSevereum) APIs() []rpc.API {
	return append(sevapi.GetAPIs(s.ApiBackend), []rpc.API{
		{
			Namespace: "sev",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "sev",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "sev",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *LightSevereum) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *LightSevereum) BlockChain() *light.LightChain      { return s.blockchain }
func (s *LightSevereum) TxPool() *light.TxPool              { return s.txPool }
func (s *LightSevereum) Engine() consensus.Engine           { return s.engine }
func (s *LightSevereum) LesVersion() int                    { return int(ClientProtocolVersions[0]) }
func (s *LightSevereum) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *LightSevereum) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *LightSevereum) Protocols() []p2p.Protocol {
	return s.makeProtocols(ClientProtocolVersions)
}

// Start implements node.Service, starting all internal goroutines needed by the
// Severeum protocol implementation.
func (s *LightSevereum) Start(srvr *p2p.Server) error {
	log.Warn("Light client mode is an experimental feature")
	s.startBloomHandlers(params.BloomBitsBlocksClient)
	s.netRPCService = sevapi.NewPublicNetAPI(srvr, s.networkId)
	// clients are searching for the first advertised protocol in the list
	protocolVersion := AdvertiseProtocolVersions[0]
	s.serverPool.start(srvr, lesTopic(s.blockchain.Genesis().Hash(), protocolVersion))
	s.protocolManager.Start(s.config.LightPeers)
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Severeum protocol.
func (s *LightSevereum) Stop() error {
	s.odr.Stop()
	s.bloomIndexer.Close()
	s.chtIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()
	s.engine.Close()

	s.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
