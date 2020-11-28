package encoder

import (
	"fmt"
	"strconv"
)

func printf(ep EncoderParameters, data []byte) ([]byte, error) {
	var x interface{} = string(data)
	if val, ok := ep["printf_int"]; !ok || val != "false" {
		if y, err := strconv.Atoi(string(data)); err != nil {
			return nil, err
		} else {
			x = y
		}
	}

	f := "0x%04x"
	if val, ok := ep["printf_fmt"]; ok {
		f = val
	}

	return []byte(fmt.Sprintf(f, x)), nil
}
