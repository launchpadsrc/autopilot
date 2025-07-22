package jsondump

import "encoding/json"

func Dump(v any) string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}

func DumpBytes(v any) []byte {
	data, _ := json.MarshalIndent(v, "", "  ")
	return data
}
