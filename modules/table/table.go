package table

import (
    "github.com/olekukonko/tablewriter"
    "strconv"
    "os"
)

func Show(data []map[string]interface{}, headers []string) {
    table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)

	for _, v := range data {
		sl := []string{}
		for _, h := range headers {
			if str, ok := v[h].(string); ok {
				sl = append(sl, str)
			}
			if i, ok := v[h].(int64); ok {
				str := strconv.Itoa(int(i))
				sl = append(sl, str)
			}
		}
		table.Append(sl)
	}
	table.Render()
}
