.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/severeum/go-severeum/cmd/swarm
	go build -o bin/ssev     github.com/severeum/go-severeum/cmd/ssev
	go build -o bin/bootnode github.com/severeum/go-severeum/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/severeum/go-severeum/swarm/...
