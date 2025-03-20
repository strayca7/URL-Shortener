package util

import (
    "encoding/json"
    "errors"
)

func ParseJSON(data []byte, v interface{}) error {
    if len(data) == 0 {
        return errors.New("empty JSON data")
    }
    return json.Unmarshal(data, v)
}