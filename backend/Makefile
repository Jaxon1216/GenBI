# GenBI Go 后端便捷命令
# 用法：make run / make consumer / make build / make vet / make test

.PHONY: run consumer build vet test tidy fmt

# 启动 HTTP 服务（:8080）
run:
	go run ./cmd/server

# 启动 RabbitMQ 消费者（异步图表生成）
consumer:
	go run ./cmd/consumer

# 编译两个二进制到 bin/
build:
	go build -o bin/genbi-server ./cmd/server
	go build -o bin/genbi-consumer ./cmd/consumer

vet:
	go vet ./...

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	gofmt -w .
