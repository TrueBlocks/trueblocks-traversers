all:
	@echo building...
	@go mod tidy
	@go build -o ./bin/accounting main.go processData.go

test:
	@go test ./...

rebuild:
	@make all
	@cd clients ; make rebuild

excel:
	@make all
	@cd clients/trueblocks ; make excel
