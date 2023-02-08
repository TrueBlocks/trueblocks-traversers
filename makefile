all:
	@echo building...
	@go mod tidy
	@go build -o ./bin/accounting main.go

test:
	@go test ./...

data:
	@cd clients ; make data
