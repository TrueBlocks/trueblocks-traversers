all:
	@echo building...
	@go mod tidy
	@go build -o ./bin/accounting main.go processData.go

test:
	@go test ./...

rebuild:
	@make all
	@cd clients ; make rebuild

update:
	@go get "github.com/TrueBlocks/trueblocks-sdk/v5@latest"
	@go get github.com/TrueBlocks/trueblocks-core/src/apps/chifra@latest

excel:
	@make all
	@cd clients/trueblocks ; make excel
