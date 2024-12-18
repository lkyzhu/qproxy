# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod 

SUFFIX=""
ifeq ($(OS), )
	OS=$(shell go env GOOS)
else ifeq($(OS), "windows")
	SUFFIX=".exe"
endif

ifeq ($(ARCH), )
	ARCH=$(shell go env GOARCH)
endif

ifeq ($(VERSION), )
	VERSION="0.0.1"
endif

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
	GOOS=$(OS) GOARCH=$(ARCH) $(GOBUILD) -o $(BUILD_PATH)/qproxy/qproxy$(SUFFIX) -v ./main.go
	tar -zcvf  $(BUILD_PATH)/qproxy-$(OS)-$(ARCH)-$(VERSION).tar.gz -C $(BUILD_PATH) qproxy --remove-files

build: init qproxy

init:
	mkdir -p $(BUILD_PATH)
	$(GOMOD) tidy

clean:
	$(GOCLEAN)
	rm -fr $(BUILD_PATH)

