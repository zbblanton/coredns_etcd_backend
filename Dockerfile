FROM golang AS builder

RUN git clone https://github.com/coredns/coredns && \
    cd coredns && \
    sed -i '/^etcd:etcd/a coredns_etcd_backend:github.com/zbblanton/coredns_etcd_backend' plugin.cfg && \
    make

FROM alpine

COPY --from=builder /go/coredns/coredns coredns

ENTRYPOINT ["/coredns"]
