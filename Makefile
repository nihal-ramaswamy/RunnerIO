run_service:
	@go run ./cmd/app/service/main.go

build_service:
	@go build -o bin/app ./cmd/app/service/main.go

run_engine:
	@go run ./cmd/app/engine/main.go

build_engine:
	@go build -o bin/app ./cmd/app/engine/main.go

test:
	@go test -v ./...

clean:
	@rm -rf bin
