FROM golang:1.19-alpine AS builder

WORKDIR /avail-settlement

RUN apk add --update-cache gcc musl-dev

COPY . .
RUN go mod download

RUN go build -o avail-settlement server/main.go
RUN go build -o p2pkeytoid docker/local/p2pkeytoid.go
RUN $(cd third_party/polygon-edge && go build -o polygon-edge .)

FROM alpine:latest AS runner

RUN apk --no-cache add ca-certificates jq bash

WORKDIR /

COPY configs /configs
COPY docker/local/genesis.json /configs
COPY docker/local/sequencer-1.yaml /configs
COPY docker/local/sequencer-2.yaml /configs
COPY docker/local/sequencer-3.yaml /configs
COPY docker/local/watchtower-1.yaml /configs
COPY docker/local/configure.sh /configure.sh

COPY --from=builder /avail-settlement/avail-settlement /avail-settlement
COPY --from=builder /avail-settlement/p2pkeytoid /p2pkeytoid
COPY --from=builder /avail-settlement/third_party/polygon-edge/polygon-edge /polygon-edge

# Configuration script generates secrets for the node & adjusts
# the genesis.json bootnode list accordingly.
RUN /configure.sh -b sequencer-1
RUN /configure.sh sequencer-2
RUN /configure.sh sequencer-3
RUN /configure.sh watchtower-1

# Expose json-rpc, libp2p and grpc ports
EXPOSE 8545 9632 1478

ENTRYPOINT ["/avail-settlement"]
