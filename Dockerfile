FROM ubuntu:20.04

MAINTAINER Josh Ellithorpe <quest@mac.com>

ENV DEBIAN_FRONTEND="noninteractive"
RUN apt-get -y update && apt-get -y upgrade
RUN apt-get -y install gcc-8 g++-8 gcc g++ libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev wget make git

RUN mkdir /build
WORKDIR /build
RUN wget https://dl.google.com/go/go1.16.3.linux-amd64.tar.gz
RUN tar zxvf go1.16.3.linux-amd64.tar.gz
RUN mv go /usr/local
RUN mkdir -p /go/bin
RUN wget https://github.com/facebook/rocksdb/archive/refs/tags/v5.18.4.tar.gz
RUN tar zxvf v5.18.4.tar.gz

ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

WORKDIR /build/rocksdb-5.18.4
RUN make CC=gcc-8 CXX=g++-8 shared_lib

ENV ROCKSDB_PATH="/build/rocksdb-5.18.4"
ENV CGO_CFLAGS="-I/$ROCKSDB_PATH/include"
ENV CGO_LDFLAGS="-L/$ROCKSDB_PATH -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd"
ENV LD_LIBRARY_PATH=$ROCKSDB_PATH

RUN mkdir /smart_bch
WORKDIR /smart_bch
RUN git clone https://github.com/smartbch/moeingevm.git
RUN git clone https://github.com/smartbch/smartbch.git

WORKDIR /smart_bch/moeingevm/evmwrap
RUN make

ENV EVMWRAP=/smart_bch/moeingevm/evmwrap/host_bridge/libevmwrap.so

WORKDIR /smart_bch/smartbch
RUN go install github.com/smartbch/smartbch/cmd/smartbchd
RUN smartbchd init smart1 --chain-id 0x1 --home /root/.smartbchd --init-balance=10000000000000000000 --test-keys="0xe3d9be2e6430a9db8291ab1853f5ec2467822b33a1a08825a22fab1425d2bff9,0x5a09e9d6be2cdc7de8f6beba300e52823493cd23357b1ca14a9c36764d600f5e"

VOLUME ["/root/.smartbchd"]

ENTRYPOINT ["smartbchd"]
EXPOSE 8545
