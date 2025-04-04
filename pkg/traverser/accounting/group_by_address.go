package accounting

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
type GroupByAddress struct {
	Opts   traverser.Options
	Source string
	Values map[string]uint64
}

func (c *GroupByAddress) Traverse(r *types.Statement) {
	if len(c.Values) == 0 {
		c.Values = make(map[string]uint64)
	}
	c.Values[c.GetKey(r)]++
}

func (c *GroupByAddress) GetKey(r *types.Statement) string {
	switch c.Source {
	case "senders":
		return c.Source + "_" + r.Sender.String()
	case "recipients":
		return c.Source + "_" + r.Recipient.String()
	case "pairings":
		fallthrough
	default:
		return c.Source + "_" + r.Sender.String() + "_" + r.Recipient.String()
	}
}

func (c *GroupByAddress) Result() string {
	return c.Name() + "\n" + c.reportValues("Addresses", c.Values)
}

func (c *GroupByAddress) Name() string {
	return colors.Green + reflect.TypeOf(c).Elem().String() + colors.Off
}

func (c *GroupByAddress) Sort(array []*types.Statement) {
	// Nothing to do
}

func (c *GroupByAddress) reportValues(msg string, m map[string]uint64) string {
	_ = msg
	type stats struct {
		Sender    base.Address
		Recipient base.Address
		Name      string
		Count     uint64
		Status    string
	}
	nTransfers := 0
	counter := 0

	arr := make([]stats, 0, len(m))
	for k, v := range m {
		parts := strings.Split(k, "_")
		record := stats{
			Count:  v,
			Status: parts[0],
			Name:   c.Opts.Names[base.HexToAddress(parts[1])].Name,
		}
		if k == "sender" {
			record.Sender = base.HexToAddress(parts[1])
			record.Name = c.Opts.Names[base.HexToAddress(parts[1])].Name
		} else if k == "recipient" {
			record.Recipient = base.HexToAddress(parts[1])
			record.Name = c.Opts.Names[base.HexToAddress(parts[1])].Name
		} else {
			record.Sender = base.HexToAddress(parts[1])
			record.Recipient = base.HexToAddress(parts[2])
			record.Name = c.Opts.Names[base.HexToAddress(parts[1])].Name + "," + c.Opts.Names[base.HexToAddress(parts[2])].Name
		}
		arr = append(arr, record)

		nTransfers += int(v)
		counter++
	}
	sort.Slice(arr, func(i, j int) bool {
		if arr[i].Count == arr[j].Count {
			v1 := arr[i].Sender.Hex() + arr[i].Recipient.Hex()
			v2 := arr[j].Sender.Hex() + arr[j].Recipient.Hex()
			if v1 == v2 {
				return arr[i].Name < arr[j].Name
			}
			return v1 < v2
		}
		return arr[i].Count > arr[j].Count
	})

	proper := strings.ToUpper(c.Source[:1]) + c.Source[1:len(c.Source)]
	ret := fmt.Sprintf("Number of %s: %d\n", proper, counter)
	ret += fmt.Sprintf("Number of Transfers: %d\n\n", nTransfers)

	source := proper[:len(proper)-1]
	ret += source + "\n"
	if c.Source == "pairings" {
		ret += "Count,Sender,Sender Name,Recipient,Recipient Name\n"
	} else {
		ret += "Count," + source + "," + source + " Name\n"
	}
	for _, val := range arr {
		if val.Status == c.Source {
			switch c.Source {
			case "senders":
				ret += fmt.Sprintf("%d,%s,%s\n", val.Count, val.Sender, val.Name)
			case "recipients":
				ret += fmt.Sprintf("%d,%s,%s\n", val.Count, val.Recipient, val.Name)
			case "pairings":
				fallthrough
			default:
				ret += fmt.Sprintf("%d,%s,%s,%s,%s\n", val.Count, val.Sender, val.Name, val.Recipient, c.Opts.Names[val.Recipient].Name)
			}
		}
	}

	return ret
}
