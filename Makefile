
ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

deps:
ifeq (, $(shell which polygon-edge))
	git submodule update --init third_party/polygon-edge
	cd third_party/polygon-edge && \
	make build && \
	mv main $(GOPATH)/bin/polygon-edge
endif

.PHONY:deps