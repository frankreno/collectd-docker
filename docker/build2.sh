#!/bin/sh -e

export GOLANG_VERSION="1.6"
export GOLANG_DOWNLOAD_URL="https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz"
export GOLANG_DOWNLOAD_SHA256="5470eac05d273c74ff8bac7bef5bad0b5abbd1c4052efbdbc8db45332e836b0b"
export GOLANG_DOWNLOAD_DESTINATION="/tmp/go${GOLANG_VERSION}.linux-amd64.tar.gz"
export GOPATH="/go"
export PATH="${GOPATH}/bin:/usr/local/go/bin:${PATH}"

cd /go/src/github.com/bobrik/collectd-docker/collector
godep restore
go get github.com/bobrik/collectd-docker/collector/...

cd /

cp /go/bin/collectd-docker-collector /usr/bin/collectd-docker-collector
cp /go/bin/reefer /usr/bin/reefer

cp /go/src/github.com/bobrik/collectd-docker/docker/collectd.conf.tpl /etc/collectd/collectd.conf.tpl
cp /go/src/github.com/bobrik/collectd-docker/docker/run.sh /run.sh

#apt-get remove -y git curl ca-certificates
#apt-get autoremove -y

rm -rf /go /usr/local/go
rm -rf /var/lib/apt/lists/*
