protoDir := $(CURDIR)/proto/

.PHONY: all
all: genProto
	go build

.PHONY: genProto
genProto:
	protoc --go_out=plugins=grpc,paths=source_relative:"$(protoDir)" greeter.proto -I "$(protoDir)"
	@echo "---------------------------------------------------------------------------------------"