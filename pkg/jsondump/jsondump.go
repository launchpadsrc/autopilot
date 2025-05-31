package jsondump

import "encoding/json"

func Dump(v any) string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}
