STAKING_CONTRACT_PATH=.$(pwd)/third_party/avail-settlement-contracts/staking/
GOOS=
GOARCH=

ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

.PHONY: protoc
protoc:
	protoc --go_out=. --go-grpc_out=. -I . ./pkg/snapshot/proto/*.proto

run-benchmarks:
	go test ./tests -bench=. -run ^$$

.PHONY: bootstrap-secrets
bootstrap-secrets: build-server
	./avail-settlement secrets init --insecure --data-dir ./data/avail-bootnode-1
	./avail-settlement secrets init --insecure --data-dir ./data/avail-node-1
	./avail-settlement secrets init --insecure --data-dir ./data/avail-node-2

build-staking-contract:
	cd $(STAKING_CONTRACT_PATH) && make build

build-server:
	GOOS=${GOOS} GOARCH=${GOARCH} go build -o avail-settlement main.go

build-assm:
	cd assm && GOOS=${GOOS} GOARCH=${GOARCH} go build

tools-wallet:
	cd tools/wallet && GOOS=${GOOS} GOARCH=${GOARCH} go build

build: build-server

build-all: build

start-bootstrap-sequencer: build
	rm -rf data/avail-bootnode-1/blockchain/
	rm -rf data/avail-bootnode-1/trie/
	./avail-settlement server --bootstrap --config-file="./configs/bootstrap-sequencer.yaml" --account-config-file="./configs/account-bootstrap-sequencer" --fraud-srv-listen-addr ":9990"

start-sequencer: build
	rm -rf data/avail-node-1/blockchain/
	rm -rf data/avail-node-1/trie/
	./avail-settlement server --config-file="./configs/sequencer-1.yaml" --account-config-file="./configs/account-sequencer" --fraud-srv-listen-addr ":9991"

start-watchtower: build
	rm -rf data/avail-watchtower-1/blockchain/
	rm -rf data/avail-watchtower-1/trie/
	./avail-settlement server --config-file="./configs/watchtower-1.yaml" --account-config-file="./configs/account-watchtower"

create-accounts: create-bootstrap-sequencer-account create-sequencer-account create-watchtower-account

create-bootstrap-sequencer-account: build-server
	./avail-settlement availaccount --balance 6 --path ./configs/account-bootstrap-sequencer
	
create-sequencer-account: build-server
	./avail-settlement availaccount --balance 6 --path ./configs/account-sequencer

create-watchtower-account: build-server
	./avail-settlement availaccount --balance 6 --path ./configs/account-watchtower
