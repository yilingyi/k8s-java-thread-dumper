IMAGE?="yilingyi/k8s-java-thread-dumper"
GOOS?=$(shell uname -s)
override GOOS:=$(shell echo ${GOOS} | tr '[A-Z]' '[a-z]')
TARGET?="jdd-${GOOS}"

build:
	echo ${GOOS}
	GOOS=${GOOS} go build -o ${TARGET} cmd/main.go
docker:
	make build GOOS=linux
	docker buildx build -t ${IMAGE} . --load
