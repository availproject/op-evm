GOOS=
GOARCH=

.PHONY: protoc
protoc:
	protoc --go_out=. --go-grpc_out=. -I . ./pkg/snapshot/proto/*.proto

.PHONY: run-benchmarks
run-benchmarks:
	go test -v=1 ./tests -bench=. -run ^$$

.PHONY: build
build:
	GOOS=${GOOS} GOARCH=${GOARCH} go build -o op-evm main.go

.PHONY: start-bootstrap-sequencer
start-bootstrap-sequencer: build
	rm -rf data/avail-bootnode-1/blockchain/
	rm -rf data/avail-bootnode-1/trie/
	./op-evm server --bootstrap --config-file="./configs/bootstrap-sequencer.yaml" --account-config-file="./data/test-accounts/account-bootstrap-sequencer" --fraud-srv-listen-addr ":9990"

.PHONY: start-sequencer
start-sequencer: build
	rm -rf data/avail-node-1/blockchain/
	rm -rf data/avail-node-1/trie/
	./op-evm server --config-file="./configs/sequencer-1.yaml" --account-config-file="./data/test-accounts/account-sequencer" --fraud-srv-listen-addr ":9991"

.PHONY: start-watchtower
start-watchtower: build
	rm -rf data/avail-watchtower-1/blockchain/
	rm -rf data/avail-watchtower-1/trie/
	./op-evm server --config-file="./configs/watchtower-1.yaml" --account-config-file="./data/test-accounts/account-watchtower"

.PHONY: create-accounts
create-accounts: create-bootstrap-sequencer-account create-sequencer-account create-watchtower-account

.PHONY: create-bootstrap-sequencer-account
create-bootstrap-sequencer-account: build
	./op-evm availaccount --balance 6 --path ./data/test-accounts/account-bootstrap-sequencer

.PHONY: create-sequencer-account
create-sequencer-account: build
	./op-evm availaccount --balance 6 --path ./data/test-accounts/account-sequencer

.PHONY: create-watchtower-account
create-watchtower-account: build
	./op-evm availaccount --balance 6 --path ./data/test-accounts/account-watchtower
