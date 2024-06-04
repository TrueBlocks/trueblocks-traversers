package traverser

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/names"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/mytypes"
)

type Traverser[T mytypes.RawType] interface {
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
	AddrFilters map[mytypes.Address]bool
	DateFilters []mytypes.DateTime
	Names       map[base.Address]types.SimpleName
}

func GetOptions(addressFn, filterFn string) Options {
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
				} else if mytypes.IsValidPeriod(a) {
					ret.Period = a
				}
			}
		}
	}

	ret.Names, _ = names.LoadNamesMap("mainnet", names.Regular|names.Custom|names.Prefund, []string{})
	ret.Names[base.HexToAddress("0x")] = types.SimpleName{Name: "Creation/Mint"}
	log.Println(colors.Yellow+"Loaded", len(ret.Names), "names...", colors.Off)

	lines := AsciiFileToLines(addressFn)
	for _, line := range lines {
		if len(line) > 0 {
			parts := strings.Split(line, ",")
			if len(parts) > 2 {
				name := types.SimpleName{}
				name.Tags = parts[0]
				name.Address = base.HexToAddress(parts[1])
				name.Name = parts[2]
				if name.Tags < "20" {
					name.IsCustom = true
				}
				ret.Names[name.Address] = name
				// log.Println(len(ret.Names), name)
				// log.Println()
			}
		}
	}
	log.Println(colors.Yellow+"Loaded", len(lines), "addresses...", colors.Off)

	lines = AsciiFileToLines(filterFn)
	if len(lines) > 0 {
		ret.AddrFilters = make(map[mytypes.Address]bool)
		ret.DateFilters = make([]mytypes.DateTime, 0, 2)
		for _, line := range lines {
			if strings.HasPrefix(line, "#") {
				continue
			} else if strings.HasPrefix(line, "0x") {
				parts := strings.Split(line, ",")
				if len(parts) != 2 {
					log.Fatal("Invalid filter line: ", line)
					os.Exit(1)
				}
				ret.AddrFilters[mytypes.Address(parts[0])] = true
			} else if strings.HasSuffix(line, "Date") {
				if len(ret.DateFilters) == 2 {
					logger.Fatal("At most two date filters are allowed:", line)
				}
				parts := strings.Split(line, ",")
				if len(parts) != 2 {
					log.Fatal("Invalid filter line: ", line)
					os.Exit(1)
				}
				dt := mytypes.DateTime{}
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

func AsciiFileToLines(filename string) []string {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		return []string{}
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var ret []string
	for scanner.Scan() {
		ret = append(ret, scanner.Text())
	}
	file.Close()
	return ret
}
