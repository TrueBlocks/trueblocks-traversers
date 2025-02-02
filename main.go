package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/file"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	sdk "github.com/TrueBlocks/trueblocks-sdk/v4"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/mytypes"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser/accounting"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser/logs"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser/stats"

	"github.com/gocarina/gocsv"
)

// --------------------------------
func main() {
	if false {
		main2()
	} else {
		rootFolder, _ := os.Getwd()
		summaryFolder := filepath.Join(rootFolder, "/summary/")
		if !file.FolderExists(summaryFolder) {
			log.Println(Usage("{0}} not found.", summaryFolder))
			os.Exit(0)
		}

		addressFn := filepath.Join(rootFolder, "addresses.csv")
		if !file.FileExists(addressFn) {
			log.Println(Usage("{0} not found.", addressFn))
			os.Exit(0)
		}

		filterFn := filepath.Join(rootFolder, "filters.csv")
		// if !FileExists(filterFn) {
		// 	log.Println(Usage("{0} not found.", filterFn))
		// 	os.Exit(0)
		// }

		opts := traverser.GetOptions(addressFn, filterFn)
		statTraversers := stats.GetTraversers(opts)
		reconTraversers := accounting.GetTraversers(opts)
		logTraversers := logs.GetTraversers(opts)

		nRecons := 0
		nLogs := 0
		filepath.Walk(summaryFolder, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			isRecon := strings.Contains(path, "all_recons.csv")
			isLog := strings.Contains(path, "all_logs.csv")

			if !isRecon && !isLog {
				return nil
			}

			var theFile *os.File
			theFile, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
			if err != nil {
				return err
			}
			defer theFile.Close()

			log.Println("Reading file", path)
			if isRecon {
				nRecons++
				recons := []*mytypes.RawReconciliation{}
				if err := gocsv.UnmarshalFile(theFile, &recons); err != nil {
					if !errors.Is(err, gocsv.ErrEmptyCSVFile) {
						logger.Error("Path:", path, err)
						return err
					}
					return nil
				}

				log.Println(colors.Yellow+"Loaded", len(recons), "recons from", path, colors.Off)
				for _, r := range recons {
					for _, a := range statTraversers {
						a.Traverse(float64(r.BlockNumber))
					}

					sorted := map[string]bool{}
					for _, a := range reconTraversers {
						if !sorted[a.Name()] {
							a.Sort(recons)
							sorted[a.Name()] = true
						}
						a.Traverse(r)
					}
				}
			} else if isLog {
				nLogs++
				logs := []*mytypes.RawLog{}
				if err := gocsv.UnmarshalFile(theFile, &logs); err != nil {
					if !errors.Is(err, gocsv.ErrEmptyCSVFile) {
						logger.Error("Path:", path, err)
						return err
					}
					return nil
				}

				log.Println(colors.Yellow+"Loaded", len(logs), "logs from", path, colors.Off)
				for _, l := range logs {
					for _, a := range logTraversers {
						a.Sort(logs)
						a.Traverse(l)
					}
				}
			} else {
				log.Panic("Should never get here")
			}

			return nil
		})

		if nRecons == 0 && nLogs == 0 {
			log.Println(Usage("No recon or log files found in {0}", summaryFolder))
			os.Exit(0)
		}

		for _, a := range statTraversers {
			fmt.Println(a.Result())
		}

		for _, a := range reconTraversers {
			fmt.Println(a.Result())
		}

		for _, a := range logTraversers {
			fmt.Println(a.Result())
		}
	}
}

func Usage(msg string, values ...string) error {
	ret := msg
	for index, val := range values {
		rep := "{" + strconv.FormatInt(int64(index), 10) + "}"
		ret = strings.Replace(ret, rep, val, -1)
	}
	return errors.New(ret)
}

func main2() {
	os.Setenv("TB_NO_USERQUERY", "true")

	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <EthereumAddress>")
	}
	address := os.Args[1]
	fmt.Printf("\n--------------------------------------------------------------------------------\n")
	fmt.Printf("Processing %s...\n", address)

	file.EstablishFolder("./raw/txs")
	file.EstablishFolder("./raw/recons")
	file.EstablishFolder("./raw/logs")
	file.EstablishFolder("./summary")
	// logger.ToggleDecoration()

	// chifra monitors --decache $1
	monitorsOptions := sdk.MonitorsOptions{
		Addrs:   []string{address},
		Globals: sdk.Globals{Decache: true},
	}
	report, _, err := monitorsOptions.Monitors()
	if err != nil {
		log.Fatalf("Error in monitors decache: %v", err)
	}
	fmt.Println(report)

	// chifra export --fmt csv --articulate --cache $1 >raw/txs/$1.csv
	exportOptions := sdk.ExportOptions{
		Addrs:      []string{address},
		Articulate: true,
		Globals: sdk.Globals{
			Cache: true,
		},
	}
	transactions, _, err := exportOptions.Export()
	if err != nil {
		log.Fatalf("Error in export transactions: %v", err)
	}
	writeTransactionsToCSV(fmt.Sprintf("raw/txs/%s.csv", address), transactions)

	// chifra export --fmt csv --articulate --accounting --statements $1 >raw/recons/$1.csv
	exportOptions.Accounting = true
	statements, _, err := exportOptions.ExportStatements()
	if err != nil {
		log.Fatalf("Error in export statements: %v", err)
	}
	writeStatementsToCSV(fmt.Sprintf("raw/recons/%s.csv", address), statements)

	// chifra export --fmt csv --articulate --logs $1 >raw/logs/$1.csv
	exportOptions.Accounting = false
	logs, _, err := exportOptions.ExportLogs()
	if err != nil {
		log.Fatalf("Error in export logs: %v", err)
	}
	writeLogsToCSV(fmt.Sprintf("raw/logs/%s.csv", address), logs)

	// Additional processing as per the original script
	processReconsFile(fmt.Sprintf("raw/recons/%s.csv", address))
	appendFilteredLogs(fmt.Sprintf("raw/logs/%s.csv", address), "summary/all_logs.csv")
	appendProcessedRecons(fmt.Sprintf("raw/recons/%s.csv", address), "summary/all_recons.csv")
}

func writeTransactionsToCSV(filename string, transactions []types.Transaction) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Cannot create file %s: %v", filename, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV headers
	headers := []string{"BlockNumber", "TxHash", "From", "To", "Value", "Gas", "GasPrice", "Nonce", "Input"}
	if err := writer.Write(headers); err != nil {
		log.Fatalf("Cannot write headers to file %s: %v", filename, err)
	}

	// Write transaction data
	for _, tx := range transactions {
		record := []string{
			fmt.Sprintf("%d", tx.BlockNumber),
			tx.Hash.Hex(),
			tx.From.Hex(),
			tx.To.Hex(),
			fmt.Sprintf("%d", tx.Value.Uint64()),
			fmt.Sprintf("%d", tx.Gas),
			fmt.Sprintf("%d", tx.GasPrice),
			fmt.Sprintf("%d", tx.Nonce),
			tx.Input,
		}
		if err := writer.Write(record); err != nil {
			log.Fatalf("Cannot write record to file %s: %v", filename, err)
		}
	}
}

func writeStatementsToCSV(filename string, statements []types.Statement) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Cannot create file %s: %v", filename, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV headers
	headers := []string{"Date", "Account", "Symbol", "Amount", "TxHash", "BlockNumber"}
	if err := writer.Write(headers); err != nil {
		log.Fatalf("Cannot write headers to file %s: %v", filename, err)
	}

	// Write statement data
	for _, stmt := range statements {
		record := []string{
			stmt.Date(),
			stmt.AccountedFor.Hex(),
			stmt.AssetSymbol,
			fmt.Sprintf("%f", stmt.AmountNet().Float64()),
			stmt.TransactionHash.Hex(),
			fmt.Sprintf("%d", stmt.BlockNumber),
		}
		if err := writer.Write(record); err != nil {
			log.Fatalf("Cannot write record to file %s: %v", filename, err)
		}
	}
}

func writeLogsToCSV(filename string, logs []types.Log) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Cannot create file %s: %v", filename, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV headers
	headers := []string{"BlockNumber", "TxHash", "Address", "Data", "Topics"}
	if err := writer.Write(headers); err != nil {
		log.Fatalf("Cannot write headers to file %s: %v", filename, err)
	}

	// Write log data
	for _, logEntry := range logs {
		topics := []string{}
		for _, topic := range logEntry.Topics {
			topics = append(topics, topic.Hex())
		}
		record := []string{
			fmt.Sprintf("%d", logEntry.BlockNumber),
			logEntry.TransactionHash.Hex(),
			logEntry.Address.Hex(),
			logEntry.Data,
			strings.Join(topics, ";"),
		}
		if err := writer.Write(record); err != nil {
			log.Fatalf("Cannot write record to file %s: %v", filename, err)
		}
	}
}

func processReconsFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Cannot open file %s: %v", filename, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Cannot read file %s: %v", filename, err)
	}

	// Extract and count unique values from column 6
	counter := make(map[string]int)
	for _, record := range records {
		if len(record) > 5 {
			value := strings.Split(record[5], "-")[0] // Extract portion before '-'
			counter[value]++
		}
	}

	// Sort and print unique counts
	var sortedKeys []string
	for key := range counter {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	fmt.Println("Unique values count from recons file:")
	for _, key := range sortedKeys {
		fmt.Printf("%d %s\n", counter[key], key)
	}
}

func appendFilteredLogs(inputFile, outputFile string) {
	input, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("Cannot open input file %s: %v", inputFile, err)
	}
	defer input.Close()

	output, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Cannot open output file %s: %v", outputFile, err)
	}
	defer output.Close()

	scanner := bufio.NewScanner(input)
	writer := bufio.NewWriter(output)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "{code:-32000 message:invalid") && !strings.Contains(line, "blockNumber") {
			_, err := writer.WriteString(line + "\n")
			if err != nil {
				log.Fatalf("Cannot write to file %s: %v", outputFile, err)
			}
		}
	}
	writer.Flush()
}

func appendProcessedRecons(inputFile, outputFile string) {
	input, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("Cannot open input file %s: %v", inputFile, err)
	}
	defer input.Close()

	output, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Cannot open output file %s: %v", outputFile, err)
	}
	defer output.Close()

	scanner := bufio.NewScanner(input)
	writer := bufio.NewWriter(output)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.ReplaceAll(line, "`", " ") // Replace backticks with spaces

		// Process CSV columns
		fields := strings.Split(line, `","`)
		if len(fields) < 24 {
			continue
		}

		// Reorder fields
		reordered := []string{
			fields[6], fields[7], fields[11], fields[5], fields[0], fields[1], fields[2], fields[3],
			fields[14], fields[21], fields[35], fields[34], fields[15], fields[16], fields[17], fields[18],
			fields[9], fields[10], fields[4], fields[8], fields[12], fields[13], fields[20], fields[19],
		}

		// Reconstruct CSV format
		outputLine := `"` + strings.Join(reordered, `","`) + `"`
		outputLine = strings.ReplaceAll(outputLine, `UTC","","`, `UTC","`) // Fixing malformed cases

		// Filter out invalid lines
		if !strings.Contains(outputLine, "{code:-32") && !strings.Contains(outputLine, "assetAddr") {
			_, err := writer.WriteString(outputLine + "\n")
			if err != nil {
				log.Fatalf("Cannot write to file %s: %v", outputFile, err)
			}
		}
	}
	writer.Flush()
}
