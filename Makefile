
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
	$(POLYGON_EDGE_BIN) secrets init --data-dir ./data/avail-bootnode-1
	$(POLYGON_EDGE_BIN) secrets init --data-dir ./data/avail-node-1
	$(POLYGON_EDGE_BIN) secrets init --data-dir ./data/avail-node-2

bootstrap-genesis:
	$(POLYGON_EDGE_BIN) genesis --dir $(POLYGON_EDGE_CONFIGS_DIR)/genesis2.json \
	--premine 0x064A4a5053F3de5eacF5E72A2E97D5F9CF55f031:1000000000000000000000 \
	--consensus ibft \
	--bootnode /ip4/127.0.0.1/tcp/10001/p2p/16Uiu2HAmMNxPzdzkNmtV97e9Y7kvHWahpGysW2Mq7GdDCDFdAcZa \
	--ibft-validator 0x1bC763b9c36Bb679B17Fc9ed01Ec5e27AF145864

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