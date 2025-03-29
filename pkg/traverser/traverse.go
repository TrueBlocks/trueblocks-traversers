package traverser

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/file"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/names"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
)

type Traversable interface {
	float64 | int64 | *types.Statement | *types.Log
}

type Traverser[T Traversable] interface {
	Traverse(t T)
	GetKey(t T) string
	Result() string
	Name() string
	Sort(array []T)
}

type Options struct {
	Period      string
	Denom       string
	Verbose     int
	AddrFilters map[base.Address]bool
	DateFilters []base.DateTime
	Names       map[base.Address]types.Name
	Accounts    map[base.Address]types.Name
}

func GetOptions() Options {
	ret := Options{}
	if len(os.Args) > 1 {
		for i, a := range os.Args {
			if i > 0 {
				if a == "--nocolor" {
					colors.ColorsOff()
				} else if a == "--verbose" {
					ret.Verbose++
				} else if a == "units" || a == "usd" || a == "wei" {
					ret.Denom = a
				} else if base.IsValidPeriod(a) {
					ret.Period = a
				}
			}
		}
	}

	ret.Names, _ = names.LoadNamesMap("mainnet", types.Regular|types.Custom|types.Prefund, []string{})
	ret.Names[base.HexToAddress("0x")] = types.Name{Name: "Creation/Mint"}
	log.Println(colors.Yellow+"Loaded", len(ret.Names), "names...", colors.Off)
	ret.Accounts = make(map[base.Address]types.Name)

	rootFolder, _ := os.Getwd()
	addressFn := filepath.Join(rootFolder, "addresses.csv")
	if !file.FileExists(addressFn) {
		log.Println(Usage("{0} not found.", addressFn))
		os.Exit(0)
	}
	lines := file.AsciiFileToLines(addressFn)
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && len(line) > 0 {
			parts := strings.Split(line, ",")
			if len(parts) > 2 {
				name := types.Name{}
				name.Tags = parts[0]
				name.Address = base.HexToAddress(parts[1])
				name.Name = parts[2]
				if name.Tags < "20" {
					name.IsCustom = true
				}
				ret.Names[name.Address] = name
				ret.Accounts[name.Address] = name
			}
		}
	}
	log.Println(colors.Yellow+"Loaded", len(lines), "addresses...", colors.Off)

	filterFn := filepath.Join(rootFolder, "filters.csv")
	if !file.FileExists(filterFn) {
		log.Println(Usage("{0} not found.", filterFn))
		os.Exit(0)
	}
	lines = file.AsciiFileToLines(filterFn)
	if len(lines) > 0 {
		ret.AddrFilters = make(map[base.Address]bool)
		ret.DateFilters = make([]base.DateTime, 0, 2)
		for _, line := range lines {
			if strings.HasPrefix(line, "#") {
				continue
			} else if strings.HasPrefix(line, "0x") {
				parts := strings.Split(line, ",")
				if len(parts) != 2 {
					log.Fatal("Invalid filter line: ", line)
					os.Exit(1)
				}
				ret.AddrFilters[base.HexToAddress(parts[0])] = true
			} else if strings.HasSuffix(line, "Date") {
				if len(ret.DateFilters) == 2 {
					logger.Fatal("At most two date filters are allowed:", line)
				}
				parts := strings.Split(line, ",")
				if len(parts) != 2 {
					log.Fatal("Invalid filter line: ", line)
					os.Exit(1)
				}
				dt := base.DateTime{}
				if err := dt.UnmarshalCSV(parts[0]); err != nil {
					logger.Fatal("invalid date filter:", line, err)
				}
				ret.DateFilters = append(ret.DateFilters, dt)
			} else {
				logger.Fatal("Invalid filter line:", line)
			}
		}
	}
	log.Println(colors.Yellow+"Loaded", len(ret.AddrFilters), "address filters...", colors.Off)
	log.Println(colors.Yellow+"Loaded", len(ret.DateFilters), "date filters...", colors.Off)

	return ret
}

func Usage(msg string, values ...string) error {
	ret := msg
	for index, val := range values {
		rep := "{" + strconv.FormatInt(int64(index), 10) + "}"
		ret = strings.Replace(ret, rep, val, -1)
	}
	return errors.New(ret)
}
