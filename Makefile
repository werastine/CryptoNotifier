.PHONY: proto clean


MODULE = github.com/werastine/CryptoNotifier

proto:
	mkdir -p pkg/pb
	protoc --go_out=pkg/pb --go_opt=module=$(MODULE)/pkg/pb \
	       --go-grpc_out=pkg/pb --go-grpc_opt=module=$(MODULE)/pkg/pb \
	       api/proto/crypto.proto