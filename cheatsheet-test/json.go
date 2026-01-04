package cheatsheettest

import "encoding/json"

func MarshalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}
