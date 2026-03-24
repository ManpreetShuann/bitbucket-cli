package output

import (
	"encoding/json"
)

type JSONFormatter struct{}

func (f *JSONFormatter) Format(data any) (string, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
