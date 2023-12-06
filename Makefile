dep:
	go mod download
dep-sync:
	go mod tidy

proto-gen:
	docker run --rm -v `pwd`/external:/defs namely/protoc-all:1.51_0  -d protos/balances -l go -o ./
	docker run --rm -v `pwd`/external:/defs namely/protoc-all:1.51_0  -d protos/orders -l go -o ./
	docker run --rm -v `pwd`/external:/defs namely/protoc-all:1.51_0  -d protos/tickets -l go -o ./
	