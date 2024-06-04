package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/mytypes"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser/accounting"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser/logs"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser/stats"

	"github.com/gocarina/gocsv"
)

// --------------------------------
func main() {
	rootFolder, _ := os.Getwd()
	summaryFolder := filepath.Join(rootFolder, "/summary/")
	if !FolderExists(summaryFolder) {
		log.Println(Usage("{0}} not found.", summaryFolder))
		os.Exit(0)
	}

	addressFn := filepath.Join(rootFolder, "addresses.csv")
	if !FileExists(addressFn) {
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

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func FolderExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func Usage(msg string, values ...string) error {
	ret := msg
	for index, val := range values {
		rep := "{" + strconv.FormatInt(int64(index), 10) + "}"
		ret = strings.Replace(ret, rep, val, -1)
	}
	return errors.New(ret)
}
