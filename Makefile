SERVICE_NAME = jwtsearchservice

lint:
	buf lint

generate:
	buf beta mod update
	buf generate --path ./proto/${SERVICE_NAME}.proto

install:
	go get google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go get github.com/bufbuild/buf/cmd/buf@latest

go_mod:
	go env -w GOPRIVATE=github.com/Cloudwalker-Technologies
	go mod tidy
	go mod download

build_project: install generate go_mod

build_run:
	cd server && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o ${SERVICE_NAME} .
	./server/${SERVICE_NAME}

build_run_hardCODE:
	cd server && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o ${SERVICE_NAME} .
	./server/${SERVICE_NAME}

build_run_silicon:
	cd server && GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-w -s" -o ${SERVICE_NAME} .
	./server/${SERVICE_NAME}