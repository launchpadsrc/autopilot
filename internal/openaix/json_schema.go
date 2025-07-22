package openaix

import "encoding/json"

type JSONSchema map[string]any

func (js JSONSchema) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any(js))
}
