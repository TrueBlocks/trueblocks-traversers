data:
	@echo making dao-01

	@mkdir -p output/recons output/logs
	@rm -fR output/**/*.csv

	@echo "Testing stats..."
	@../../bin/accounting stats.counter total average max min --nocolor >output/stats.csv 2>/dev/null

	@echo "Testing counters..."
	@../../bin/accounting by_asset --nocolor >output/recons/by_asset.csv 2>/dev/null
	@../../bin/accounting by_function --nocolor >output/recons/by_function.csv 2>/dev/null

	@echo "Testing groupers..."
	@../../bin/accounting by_priced --nocolor >output/recons/by_priced.csv 2>/dev/null
	@../../bin/accounting by_address --nocolor >output/recons/by_address.csv 2>/dev/null
	@../../bin/accounting statements --nocolor usd >output/recons/recons.csv 2>/dev/null
	@../../bin/accounting senders --nocolor usd >output/recons/senders.csv 2>/dev/null
	@../../bin/accounting recipients --nocolor usd >output/recons/recipients.csv 2>/dev/null
	@../../bin/accounting pairings --nocolor usd >output/recons/pairings.csv 2>/dev/null

	@echo "Token summaries..."
	@../../bin/accounting topic_only --nocolor >output/logs/topic_only.csv 2>/dev/null
	@../../bin/accounting contract_only --nocolor >output/logs/contract_only.csv 2>/dev/null
	@../../bin/accounting contract_last --nocolor >output/logs/contract_last.csv 2>/dev/null
	@../../bin/accounting contract_first --nocolor >output/logs/contract_first.csv 2>/dev/null
	@../../bin/accounting extract --nocolor >output/logs/logs.csv 2>/dev/null

	@echo "Exporting..."
	@../../bin/accounting profit_and_loss daily units --verbose --verbose --nocolor >output/recons/daiy_p_and_l.csv
	@../../bin/accounting profit_and_loss monthly units --verbose --verbose --nocolor >output/recons/monthly_p_and_l.csv
	@make excel

excel:
	@echo "Creating spreadsheet..."
	@../../bin/accounting excel ; mv Book1.xlsx output/dao-01.xlsx

clean:
	@rm -fR output raw summary
	@mkdir -p output/logs output/recons raw/txs raw/recons raw/logs summary
