STAKING_CONTRACT_PATH=.$(pwd)/third_party/avail-settlement-contracts/staking/
GOOS=
GOARCH=

.PHONY: protoc
protoc:
	protoc --go_out=. --go-grpc_out=. -I . ./pkg/snapshot/proto/*.proto

.PHONY: run-benchmarks
run-benchmarks:
	go test ./tests -bench=. -run ^$$

.PHONY: bootstrap-secrets
bootstrap-secrets: build
	./avail-settlement secrets init --insecure --data-dir ./data/avail-bootnode-1
	./avail-settlement secrets init --insecure --data-dir ./data/avail-node-1
	./avail-settlement secrets init --insecure --data-dir ./data/avail-node-2

.PHONY: build-staking-contract
build-staking-contract:
	cd $(STAKING_CONTRACT_PATH) && make build

.PHONY: build
build:
	GOOS=${GOOS} GOARCH=${GOARCH} go build -o avail-settlement main.go

.PHONY: start-bootstrap-sequencer
start-bootstrap-sequencer: build
	rm -rf data/avail-bootnode-1/blockchain/
	rm -rf data/avail-bootnode-1/trie/
	./avail-settlement server --bootstrap --config-file="./configs/bootstrap-sequencer.yaml" --account-config-file="./configs/account-bootstrap-sequencer" --fraud-srv-listen-addr ":9990"

.PHONY: start-sequencer
start-sequencer: build
	rm -rf data/avail-node-1/blockchain/
	rm -rf data/avail-node-1/trie/
	./avail-settlement server --config-file="./configs/sequencer-1.yaml" --account-config-file="./configs/account-sequencer" --fraud-srv-listen-addr ":9991"

.PHONY: start-watchtower
start-watchtower: build
	rm -rf data/avail-watchtower-1/blockchain/
	rm -rf data/avail-watchtower-1/trie/
	./avail-settlement server --config-file="./configs/watchtower-1.yaml" --account-config-file="./configs/account-watchtower"

.PHONY: create-accounts
create-accounts: create-bootstrap-sequencer-account create-sequencer-account create-watchtower-account

.PHONY: create-bootstrap-sequencer-account
create-bootstrap-sequencer-account: build
	./avail-settlement availaccount --balance 6 --path ./configs/account-bootstrap-sequencer

.PHONY: create-sequencer-account
create-sequencer-account: build
	./avail-settlement availaccount --balance 6 --path ./configs/account-sequencer

.PHONY: create-watchtower-account
create-watchtower-account: build
	./avail-settlement availaccount --balance 6 --path ./configs/account-watchtower
