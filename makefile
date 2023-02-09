all:
	@echo building...
	@go mod tidy
	@go build -o ./bin/accounting main.go

test:
	@go test ./...

data:
	@cd clients ; make data

person-01:
	@cd clients/person-01 ; make excel

person-03:
	@cd clients/person-03 ; make excel
