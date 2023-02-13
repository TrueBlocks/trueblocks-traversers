all:
	@echo building...
	@go mod tidy
	@go build -o ./bin/accounting main.go

test:
	@go test ./...

data:
	@cd clients ; make data

rebuild:
	@go build -o ./bin/accounting main.go
	@cd clients ; make rebuild

person-01:
	@cd clients/person-01 ; make excel

person-02:
	@cd clients/person-02 ; make excel

person-03:
	@cd clients/person-03 ; make excel

trueblocks:
	@cd clients/trueblocks-wallets ; make excel

dao:
#	@cd clients/dao-01 ; ./doAll.sh ; cd -
	@cd clients/dao-01 ; ./doAll.sh ; make data ; cd -
#	@cd clients/dao-01 ; make excel ; cd -
