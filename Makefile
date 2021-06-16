# Makefile
IMAGE_NAME=quay.io/openshift-logging/log-stressor:latest
all: deploy

deploy:
	./deploy_to_openshift.sh

lint:
	go get -u golang.org/x/lint/golint
	golint go/*

build: lint
	go env -w GO111MODULE=auto
	rm -f log-stressor.zip
	rm -f log-stressor
	rm -f check-logs-sequence.zip
	rm -f check-logs-sequence
	go get -u github.com/papertrail/go-tail
	go get -u github.com/sirupsen/logrus
	go build -ldflags "-s -w" go/log-stressor/log-stressor.go
	go build -ldflags "-s -w" go/check-logs-sequence/check-logs-sequence.go

image:
	podman build -t $(IMAGE_NAME) .

test: build
	./check-logs-sequence -c 1 -l info -s 0 -f go/check-logs-sequence/example.stress.log
