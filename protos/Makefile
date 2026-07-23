PROTOC_VERSION := 35.1
PROTOC_GEN_GO_VERSION := v1.36.11
PROTOC_GEN_GO_GRPC_VERSION := v1.6.2
PROTOC_GEN_GRPC_JAVA_VERSION := 1.69.1

PROTO_DIR := .
PROTO_FILE := $(PROTO_DIR)/iwf.proto

SERVER_GO_OUT := ../server/gen/iwfpb
SDK_GO_OUT := ../sdk-go/gen/iwfpb
SDK_JAVA_OUT := ../sdk-java/src/main/java
SDK_PYTHON_OUT := ../sdk-python/iwf/iwfpb

SERVER_GO_PKG := github.com/superdurable/iwf/gen/iwfpb
SDK_GO_PKG := github.com/superdurable/iwf/sdk-go/gen/iwfpb

GOPATH_BIN := $(shell go env GOPATH)/bin
export PATH := $(GOPATH_BIN):$(PATH)

TOOLS_DIR := .tools
GRPC_JAVA_PLUGIN := $(TOOLS_DIR)/protoc-gen-grpc-java
PYTHON_VENV := $(TOOLS_DIR)/pyvenv
PYTHON := $(PYTHON_VENV)/bin/python
PIP := $(PYTHON_VENV)/bin/pip

UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_S),Darwin)
  ifeq ($(UNAME_M),arm64)
    GRPC_JAVA_CLASSIFIER := osx-aarch_64
  else
    GRPC_JAVA_CLASSIFIER := osx-x86_64
  endif
else ifeq ($(UNAME_S),Linux)
  ifeq ($(UNAME_M),aarch64)
    GRPC_JAVA_CLASSIFIER := linux-aarch_64
  else
    GRPC_JAVA_CLASSIFIER := linux-x86_64
  endif
else
  GRPC_JAVA_CLASSIFIER := windows-x86_64
endif

.PHONY: proto proto-go proto-java proto-python clean ensure-tools

proto: proto-go proto-java proto-python

proto-go: ensure-tools
	rm -rf $(SERVER_GO_OUT) $(SDK_GO_OUT)
	mkdir -p $(SERVER_GO_OUT) $(SDK_GO_OUT)
	protoc \
		-I $(PROTO_DIR) \
		--go_out=$(SERVER_GO_OUT) \
		--go_opt=paths=source_relative \
		--go_opt=Miwf.proto=$(SERVER_GO_PKG) \
		--go-grpc_out=$(SERVER_GO_OUT) \
		--go-grpc_opt=paths=source_relative \
		--go-grpc_opt=Miwf.proto=$(SERVER_GO_PKG) \
		$(PROTO_FILE)
	protoc \
		-I $(PROTO_DIR) \
		--go_out=$(SDK_GO_OUT) \
		--go_opt=paths=source_relative \
		--go_opt=Miwf.proto=$(SDK_GO_PKG) \
		--go-grpc_out=$(SDK_GO_OUT) \
		--go-grpc_opt=paths=source_relative \
		--go-grpc_opt=Miwf.proto=$(SDK_GO_PKG) \
		$(PROTO_FILE)

proto-java: ensure-tools $(GRPC_JAVA_PLUGIN)
	rm -rf $(SDK_JAVA_OUT)/io/superdurable/gen
	mkdir -p $(SDK_JAVA_OUT)
	protoc \
		-I $(PROTO_DIR) \
		--java_out=$(SDK_JAVA_OUT) \
		--plugin=protoc-gen-grpc-java=$(GRPC_JAVA_PLUGIN) \
		--grpc-java_out=$(SDK_JAVA_OUT) \
		$(PROTO_FILE)

proto-python: ensure-tools
	rm -rf $(SDK_PYTHON_OUT)
	mkdir -p $(SDK_PYTHON_OUT)
	$(PYTHON) -m grpc_tools.protoc \
		-I $(PROTO_DIR) \
		--python_out=$(SDK_PYTHON_OUT) \
		--grpc_python_out=$(SDK_PYTHON_OUT) \
		--pyi_out=$(SDK_PYTHON_OUT) \
		$(PROTO_FILE)
	printf '%s\n' '# Generated package for iWF protobuf stubs.' > $(SDK_PYTHON_OUT)/__init__.py
	# grpc stubs import "iwf_pb2"; fix to relative for package layout.
	@if [ -f $(SDK_PYTHON_OUT)/iwf_pb2_grpc.py ]; then \
		sed -i.bak 's/^import iwf_pb2 as/from . import iwf_pb2 as/' $(SDK_PYTHON_OUT)/iwf_pb2_grpc.py && rm -f $(SDK_PYTHON_OUT)/iwf_pb2_grpc.py.bak; \
	fi

clean:
	rm -rf $(SERVER_GO_OUT) $(SDK_GO_OUT)
	rm -rf $(SDK_JAVA_OUT)/io/superdurable/gen
	rm -rf $(SDK_PYTHON_OUT)

ensure-tools: $(PYTHON_VENV)/.grpcio-tools
	@which protoc > /dev/null 2>&1 || (echo "protoc not found. Install protoc $(PROTOC_VERSION)" && exit 1)
	@which protoc-gen-go > /dev/null 2>&1 || go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	@which protoc-gen-go-grpc > /dev/null 2>&1 || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)

$(PYTHON_VENV)/.grpcio-tools:
	mkdir -p $(TOOLS_DIR)
	python3 -m venv $(PYTHON_VENV)
	$(PIP) install --upgrade pip
	$(PIP) install 'grpcio-tools>=1.60'
	touch $@

$(GRPC_JAVA_PLUGIN):
	mkdir -p $(TOOLS_DIR)
	curl -fsSL -o $(GRPC_JAVA_PLUGIN) \
		"https://repo1.maven.org/maven2/io/grpc/protoc-gen-grpc-java/$(PROTOC_GEN_GRPC_JAVA_VERSION)/protoc-gen-grpc-java-$(PROTOC_GEN_GRPC_JAVA_VERSION)-$(GRPC_JAVA_CLASSIFIER).exe"
	chmod +x $(GRPC_JAVA_PLUGIN)
