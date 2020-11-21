package encoder

import (
	"bytes"
	"encoding/json"
)

func jsonString(ep EncoderParameters, data []byte) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := json.NewEncoder(buf)

	if val, ok := ep["html_escape"]; ok && val == "true" {
		enc.SetEscapeHTML(true)
	} else {
		enc.SetEscapeHTML(false)
	}

	if err := enc.Encode(string(data)); err != nil {
		return nil, err
	}

	result := buf.Bytes()
	return result[1 : len(result)-2], nil
}
