POLYGON_EDGE_BIN=.$(pwd)/third_party/polygon-edge/polygon-edge
POLYGON_EDGE_DATA_DIR=$(pwd)/data
POLYGON_EDGE_CONFIGS_DIR=$(shell pwd)/configs
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

install-polygon-edge:
	go install github.com/0xPolygon/polygon-edge@v0.8.1

run-benchmarks:
	go test ./tests -bench=. -run ^$$

bootstrap-config:
	$(POLYGON_EDGE_BIN) server export --type yaml
	mv default-config.yaml configs/edge-config.yaml
	sed -i 's/genesis.json/configs\/genesis.json/g' configs/edge-config.yaml
	sed -i 's/log_level: INFO/log_level: DEBUG/g' configs/edge-config.yaml
	sed -i 's/data_dir: ""/data_dir: ".\/data\/avail-chain-1"/g' configs/edge-config.yaml

bootstrap-secrets: build-server
	./avail-settlement secrets init --insecure --data-dir ./data/avail-bootnode-1
	./avail-settlement secrets init --insecure --data-dir ./data/avail-node-1
	./avail-settlement secrets init --insecure --data-dir ./data/avail-node-2

bootstrap-genesis:
	rm $(POLYGON_EDGE_CONFIGS_DIR)/genesis2.json || true
	$(POLYGON_EDGE_BIN) genesis --dir $(POLYGON_EDGE_CONFIGS_DIR)/genesis2.json \
	--name polygon-avail-settlement \
	--premine 0x064A4a5053F3de5eacF5E72A2E97D5F9CF55f031:1000000000000000000000 \
	--consensus ibft \
	--bootnode /ip4/127.0.0.1/tcp/10001/p2p/16Uiu2HAmMNxPzdzkNmtV97e9Y7kvHWahpGysW2Mq7GdDCDFdAcZa \
	--ibft-validator 0x1bC763b9c36Bb679B17Fc9ed01Ec5e27AF145864 \
	--ibft-validator-type "ecdsa"

build-staking-contract:
	cd $(STAKING_CONTRACT_PATH) && make build

bootstrap-staking-contract: build-staking-contract
	$(POLYGON_EDGE_BIN) genesis predeploy --chain $(POLYGON_EDGE_CONFIGS_DIR)/genesis.json \
	--predeploy-address "0x0110000000000000000000000000000000000001" \
	--artifacts-path "$(STAKING_CONTRACT_PATH)/artifacts/contracts/Staking.sol/Staking.json" \
	--constructor-args "1" \
	--constructor-args "10"
	sed -i '' -e 's/"balance": "0x0"/"balance": "0x3635c9adc5dea00000"/g' configs/genesis.json

bootstrap: bootstrap-config bootstrap-secrets bootstrap-genesis

build-fraud-contract:
	solc --abi tools/fraud/contract/Fraud.sol -o tools/fraud/contract/ --overwrite
	solc --bin tools/fraud/contract/Fraud.sol -o tools/fraud/contract/ --overwrite
	abigen --bin=./tools/fraud/contract/Contract.bin --abi=./tools/fraud/contract/Contract.abi --pkg=fraud --out=./tools/fraud/contract/Fraud.go

build-server:
	GOOS=${GOOS} GOARCH=${GOARCH} go build -o avail-settlement main.go

build-client:
	cd client && GOOS=${GOOS} GOARCH=${GOARCH} go build -o client

build-e2e:
	cd tools/e2e && GOOS=${GOOS} GOARCH=${GOARCH} go build -o e2e

build-fraud: build-fraud-contract
	cd tools/fraud && GOOS=${GOOS} GOARCH=${GOARCH} go build -o fraud

build-staking:
	cd tools/staking && GOOS=${GOOS} GOARCH=${GOARCH} go build -o staking

build-contract:
	solc --abi contracts/SetGet/SetGet.sol -o contracts/SetGet/ --overwrite
	solc --bin contracts/SetGet/SetGet.sol -o contracts/SetGet/ --overwrite
	abigen --bin=./contracts/SetGet/SetGet.bin --abi=./contracts/SetGet/SetGet.abi --pkg=setget --out=./contracts/SetGet/SetGet.go

build-assm:
	cd assm && GOOS=${GOOS} GOARCH=${GOARCH} go build

tools-wallet:
	cd tools/wallet && GOOS=${GOOS} GOARCH=${GOARCH} go build

build-tools: build-staking build-e2e

build: build-server build-client

build-all: build build-tools

start-bootstrap-sequencer: build
	rm -rf data/avail-bootnode-1/blockchain/
	rm -rf data/avail-bootnode-1/trie/
	./avail-settlement -bootstrap -config-file="./configs/bootstrap-sequencer.yaml" -account-config-file="./configs/account-bootstrap-sequencer"

start-sequencer: build
	rm -rf data/avail-node-1/blockchain/
	rm -rf data/avail-node-1/trie/
	./avail-settlement -config-file="./configs/sequencer-1.yaml" -account-config-file="./configs/account-sequencer"

start-watchtower: build
	rm -rf data/avail-watchtower-1/blockchain/
	rm -rf data/avail-watchtower-1/trie/
	./avail-settlement -config-file="./configs/watchtower-1.yaml" -account-config-file="./configs/account-watchtower"

start-e2e: build-e2e
	./tools/e2e/e2e

start-fraud: build-fraud
	./tools/fraud/fraud

start-staking: build-staking 
	./tools/staking/staking

create-accounts: create-bootstrap-sequencer-account create-sequencer-account create-watchtower-account

create-bootstrap-sequencer-account: build-server
	./avail-settlement availaccount -balance 6 -path ./configs/account-bootstrap-sequencer
	
create-sequencer-account: build-server
	./avail-settlement availaccount -balance 6 -path ./configs/account-sequencer

create-watchtower-account: build-server
	./avail-settlement availaccount -balance 6 -path ./configs/account-watchtower

deps:
ifeq (, $(shell which $(POLYGON_EDGE_BIN)))
	git submodule update --init third_party/polygon-edge
	cd third_party/polygon-edge && \
	make build && \
	mv main $(POLYGON_EDGE_BIN)
endif
	yarn install
	sh ./scripts/install_solc.sh

.PHONY: deps bootstrap
