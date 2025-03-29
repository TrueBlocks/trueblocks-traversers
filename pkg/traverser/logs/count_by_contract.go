package logs

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
)

// --------------------------------
type CountByContract struct {
	Opts   traverser.Options
	Mode   string
	Values map[string]uint64
}

func (c *CountByContract) Traverse(r *types.Log) {
	if len(c.Values) == 0 {
		c.Values = make(map[string]uint64)
	}
	c.Values[c.GetKey(r)]++
}

func (c *CountByContract) GetKey(r *types.Log) string {
	name := c.Opts.Names[r.Address].Name
	if name == "" {
		name = "Unknown"
	}
	funcName := strings.Split(strings.Replace(strings.Replace(r.CompressedLog(), "{name:", "", -1), "}", "", -1), "|")[0]
	if funcName == "" {
		funcName = "Unknown"
	}

	topic0 := "no topic"
	if len(r.Topics) > 0 {
		topic0 = r.Topics[0].Hex()
	}
	switch c.Mode {
	case "topic_only":
		return topic0 + "_" + funcName
	case "contract_only":
		return r.Address.Hex() + "_" + name
	case "contract_last":
		return topic0 + "_" + funcName + "_" + r.Address.Hex() + "_" + name
	case "contract_first":
		fallthrough
	default:
		return r.Address.Hex() + "_" + name + "_" + topic0 + "_" + funcName
	}
}

func (c *CountByContract) Result() string {
	return c.Name() + "\n" + c.reportValues("TopicsPerContract", c.Values)
}

func (c *CountByContract) Name() string {
	return colors.Green + reflect.TypeOf(c).Elem().String() + colors.Off
}

func (c *CountByContract) Sort(array []*types.Log) {
	// Nothing to do
}

func (c *CountByContract) reportValues(msg string, m map[string]uint64) string {
	type stats struct {
		Contract base.Address
		Name     string
		Topic    base.Topic
		Function string
		Count    int64
		Sort     string
	}
	nRecords := uint64(0)

	sortStr := func(a, b, c, d string) string {
		return a + "_" + b + "_" + c + "_" + d
	}

	arr := make([]stats, 0, len(m))
	for k, v := range m {
		nRecords += v
		parts := strings.Split(k, "_")
		var val stats
		switch c.Mode {
		case "topic_only":
			val = stats{Count: int64(v), Topic: parts[0], Function: parts[1]}
			val.Sort = sortStr(val.Topic, val.Function, "", "")
		case "contract_only":
			val = stats{Count: int64(v), Contract: base.HexToAddress(parts[0]), Name: parts[1]}
			val.Sort = sortStr(val.Contract.Hex(), val.Name, "", "")
		case "contract_last":
			val = stats{Count: int64(v), Topic: parts[0], Function: parts[1], Contract: base.HexToAddress(parts[2]), Name: parts[3]}
			val.Sort = sortStr(val.Topic, val.Function, val.Contract.Hex(), val.Name)
		case "contract_first":
			fallthrough
		default:
			val = stats{Count: int64(v), Contract: base.HexToAddress(parts[0]), Name: parts[1], Topic: parts[2], Function: parts[3]}
			val.Sort = sortStr(val.Contract.Hex(), val.Name, val.Topic, val.Function)
		}
		arr = append(arr, val)
	}

	sort.Slice(arr, func(i, j int) bool {
		if arr[i].Count == arr[j].Count {
			return arr[i].Sort < arr[j].Sort
		}
		return arr[i].Count > arr[j].Count
	})

	ret := fmt.Sprintf("Number of %s: %d\n", msg, len(c.Values))
	ret += fmt.Sprintf("Number of Topics: %d\n", nRecords)

	for i, val := range arr {
		switch c.Mode {
		case "topic_only":
			if i == 0 {
				ret += "Count,Topic,FuncName\n"
			}
			ret += fmt.Sprintf("%d,%s,%s\n", val.Count, val.Topic, val.Function)
		case "contract_only":
			if i == 0 {
				ret += "Count,Contract,Name\n"
			}
			ret += fmt.Sprintf("%d,%s,%s\n", val.Count, val.Contract, val.Name)
		case "contract_last":
			if i == 0 {
				ret += "Count,Topic,FuncName,Contract,Name\n"
			}
			ret += fmt.Sprintf("%d,%s,%s,%s,%s\n", val.Count, val.Topic, val.Function, val.Contract, val.Name)
		case "contract_first":
			fallthrough
		default:
			if i == 0 {
				ret += "Count,Contract,Name,Topic,FuncName\n"
			}
			ret += fmt.Sprintf("%d,%s,%s,%s,%s\n", val.Count, val.Contract, val.Name, val.Topic, val.Function)
		}
	}

	return ret
}
