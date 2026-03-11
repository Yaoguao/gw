# Simple version Makefile for protoc

PROTO_DIR := proto
GO_OUT_DIR := gen/go

.PHONY: generate
generate:
	protoc -I $(PROTO_DIR) \
		$(PROTO_DIR)/exchange/exchange.proto \
		--go_out=$(GO_OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT_DIR) --go-grpc_opt=paths=source_relative

.PHONY: clean
clean:
	rm -rf $(GO_OUT_DIR)

.PHONY: rebuild
rebuild: clean generate