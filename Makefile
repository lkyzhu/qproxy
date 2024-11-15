# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod 

#if not set release_dir
ifeq ("xx$(RELEASE_DIR)", "xx")
	RELEASE_DIR=$(ROOT_DIR)/release
endif

BUILD_PATH=$(RELEASE_DIR)
#BUILD_TARGET=
ROOT_DIR=$(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

all: build

.PHONY:qproxy
qproxy:
	mkdir -p $(BUILD_PATH)/qproxy
	$(GOBUILD) -o $(BUILD_PATH)/qproxy/qproxy -v ./main.go
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_PATH)/qproxy/qproxy.exe -v ./main.go

build: init qproxy

init:
	mkdir -p $(BUILD_PATH)
	$(GOMOD) tidy

clean:
	$(GOCLEAN)
	rm -fr $(BUILD_PATH)

