dep:
	go mod download
dep-sync:
	go mod tidy

proto-gen:
	docker run --rm -v `pwd`/external:/defs namely/protoc-all:1.51_0 -i protos -f balances/Common.proto -l go -o ./
	docker run --rm -v `pwd`/external:/defs namely/protoc-all:1.51_0 -i protos -f orders/common.proto -l go -o ./