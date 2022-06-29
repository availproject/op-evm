
POLYGON_EDGE_BIN=$(GOPATH)/bin/polygon-edge
POLYGON_EDGE_DATA_DIR=$(pwd)/data
POLYGON_EDGE_CONFIGS_DIR=$(shell pwd)/configs

ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

bootstrap-config:
	$(POLYGON_EDGE_BIN) server export --type yaml
	mv default-config.yaml configs/edge-config.yaml
	sed -i 's/genesis.json/configs\/genesis.json/g' configs/edge-config.yaml
	sed -i 's/log_level: INFO/log_level: DEBUG/g' configs/edge-config.yaml
	sed -i 's/data_dir: ""/data_dir: ".\/data\/avail-chain-1"/g' configs/edge-config.yaml

bootstrap-secrets:
	$(POLYGON_EDGE_BIN) secrets init --data-dir ./data/bootnode-1
	$(POLYGON_EDGE_BIN) secrets init --data-dir ./data/node-1
	$(POLYGON_EDGE_BIN) secrets init --data-dir ./data/node-2

bootstrap-genesis:
	$(POLYGON_EDGE_BIN) genesis --dir $(POLYGON_EDGE_CONFIGS_DIR)/genesis.json --consensus ibft --ibft-validators-prefix-path test-chain- --bootnode /ip4/127.0.0.1/tcp/10001/p2p/16Uiu2HAmLRftAwcbtdhkVHkP9N81vhKUHR5X7yigEhdvB87whpMX

bootstrap: bootstrap-config bootstrap-secrets bootstrap-genesis


build-server:
	cd server && go build -o server

build-client:
	cd client && go build -o client

build: build-server build-client

deps:
ifeq (, $(shell which polygon-edge))
	git submodule update --init third_party/polygon-edge
	cd third_party/polygon-edge && \
	make build && \
	mv main $(POLYGON_EDGE_BIN)
endif

.PHONY: deps bootstrap