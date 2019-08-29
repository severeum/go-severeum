FROM golang:1.11-alpine as builder

ARG VERSION

RUN apk add --update git gcc g++ linux-headers
RUN mkdir -p $GOPATH/src/github.com/severeum && \
    cd $GOPATH/src/github.com/severeum && \
    git clone https://github.com/seversphere/go-severeum && \
    cd $GOPATH/src/github.com/severeum/go-severeum && \
    git checkout ${VERSION} && \
    go install -ldflags "-X main.gitCommit=${VERSION}" ./cmd/swarm && \
    go install -ldflags "-X main.gitCommit=${VERSION}" ./cmd/swarm/swarm-smoke && \
    go install -ldflags "-X main.gitCommit=${VERSION}" ./cmd/ssev && \
    cp $GOPATH/bin/swarm /swarm && cp $GOPATH/bin/ssev /ssev && cp $GOPATH/bin/swarm-smoke /swarm-smoke


# Release image with the required binaries and scripts
FROM alpine:3.8
WORKDIR /
COPY --from=builder /swarm /ssev /swarm-smoke /
ADD run.sh /run.sh
ADD run-smoke.sh /run-smoke.sh
ENTRYPOINT ["/run.sh"]
