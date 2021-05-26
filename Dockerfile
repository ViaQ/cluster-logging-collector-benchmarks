FROM docker.io/amd64/golang:1.14 AS builder

WORKDIR /go/src/github.com/viaq/cluster-logging-collector-benchmarks

COPY go go
copy Makefile .

RUN make build

#@follow_tag(openshift-ose-base:ubi8)
FROM docker.io/centos:8 AS centos

RUN mkdir -p /var/log/containers /opt/app-root/src
COPY --from=builder /go/src/github.com/viaq/cluster-logging-collector-benchmarks/log-stressor /usr/bin/
COPY --from=builder /go/src/github.com/viaq/cluster-logging-collector-benchmarks/check-logs-sequence /usr/bin/

COPY conf/stressor/run-log-stressor.sh /usr/bin/run-log-stressor


WORKDIR /opt/app-root/src
ENTRYPOINT ["run-log-stressor"]

