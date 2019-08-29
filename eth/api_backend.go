// Copyright 2015 The go-severeum Authors
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

package sev

import (
	"context"
	"math/big"

	"github.com/severeum/go-severeum/accounts"
	"github.com/severeum/go-severeum/common"
	"github.com/severeum/go-severeum/common/math"
	"github.com/severeum/go-severeum/core"
	"github.com/severeum/go-severeum/core/bloombits"
	"github.com/severeum/go-severeum/core/state"
	"github.com/severeum/go-severeum/core/types"
	"github.com/severeum/go-severeum/core/vm"
	"github.com/severeum/go-severeum/sev/downloader"
	"github.com/severeum/go-severeum/sev/gasprice"
	"github.com/severeum/go-severeum/sevdb"
	"github.com/severeum/go-severeum/event"
	"github.com/severeum/go-severeum/params"
	"github.com/severeum/go-severeum/rpc"
)

// SevAPIBackend implements sevapi.Backend for full nodes
type SevAPIBackend struct {
	sev *Severeum
	gpo *gasprice.Oracle
}

// ChainConfig returns the active chain configuration.
func (b *SevAPIBackend) ChainConfig() *params.ChainConfig {
	return b.sev.chainConfig
}

func (b *SevAPIBackend) CurrentBlock() *types.Block {
	return b.sev.blockchain.CurrentBlock()
}

func (b *SevAPIBackend) SetHead(number uint64) {
	b.sev.protocolManager.downloader.Cancel()
	b.sev.blockchain.SetHead(number)
}

func (b *SevAPIBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.sev.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.sev.blockchain.CurrentBlock().Header(), nil
	}
	return b.sev.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *SevAPIBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return b.sev.blockchain.GetHeaderByHash(hash), nil
}

func (b *SevAPIBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.sev.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.sev.blockchain.CurrentBlock(), nil
	}
	return b.sev.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *SevAPIBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.sev.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.sev.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *SevAPIBackend) GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return b.sev.blockchain.GetBlockByHash(hash), nil
}

func (b *SevAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	return b.sev.blockchain.GetReceiptsByHash(hash), nil
}

func (b *SevAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	receipts := b.sev.blockchain.GetReceiptsByHash(hash)
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *SevAPIBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.sev.blockchain.GetTdByHash(blockHash)
}

func (b *SevAPIBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.sev.BlockChain(), nil)
	return vm.NewEVM(context, state, b.sev.chainConfig, *b.sev.blockchain.GetVMConfig()), vmError, nil
}

func (b *SevAPIBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.sev.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *SevAPIBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.sev.BlockChain().SubscribeChainEvent(ch)
}

func (b *SevAPIBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.sev.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *SevAPIBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.sev.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *SevAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.sev.BlockChain().SubscribeLogsEvent(ch)
}

func (b *SevAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.sev.txPool.AddLocal(signedTx)
}

func (b *SevAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.sev.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *SevAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.sev.txPool.Get(hash)
}

func (b *SevAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.sev.txPool.State().GetNonce(addr), nil
}

func (b *SevAPIBackend) Stats() (pending int, queued int) {
	return b.sev.txPool.Stats()
}

func (b *SevAPIBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.sev.TxPool().Content()
}

func (b *SevAPIBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return b.sev.TxPool().SubscribeNewTxsEvent(ch)
}

func (b *SevAPIBackend) Downloader() *downloader.Downloader {
	return b.sev.Downloader()
}

func (b *SevAPIBackend) ProtocolVersion() int {
	return b.sev.SevVersion()
}

func (b *SevAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *SevAPIBackend) ChainDb() sevdb.Database {
	return b.sev.ChainDb()
}

func (b *SevAPIBackend) EventMux() *event.TypeMux {
	return b.sev.EventMux()
}

func (b *SevAPIBackend) AccountManager() *accounts.Manager {
	return b.sev.AccountManager()
}

func (b *SevAPIBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.sev.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *SevAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.sev.bloomRequests)
	}
}
