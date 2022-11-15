export PATH := $(GOPATH)/bin
BINARY_NAME=getsetgo

build:
	go build -o ${BINARY_NAME} main.go

install:
	go build -o ${PATH}/${BINARY_NAME} main.go