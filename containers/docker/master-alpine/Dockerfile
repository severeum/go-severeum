FROM alpine:3.7

RUN \
  apk add --update go git make gcc musl-dev linux-headers ca-certificates && \
  git clone --depth 1 --branch release/1.8 https://github.com/severeum/go-severeum && \
  (cd go-severeum && make seth) && \
  cp go-severeum/build/bin/seth /seth && \
  apk del go git make gcc musl-dev linux-headers && \
  rm -rf /go-severeum && rm -rf /var/cache/apk/*

EXPOSE 8545
EXPOSE 30303

ENTRYPOINT ["/seth"]
