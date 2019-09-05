package stats

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type StatStorage interface {
	Send([]Stat) error
}

type Stat struct {
	Key    string
	Value  float64
	Labels map[string]string
}

func FromJSON(rawjson []byte) (error, []Stat) {
	var data interface{}
	err := json.Unmarshal(rawjson, &data)
	if err != nil {
		return err, nil
	}
	ret := FlattenJSON(data, "")
	return nil, ret
}

func FlattenJSON(data interface{}, key string) []Stat {
	ret := make([]Stat, 0)
	switch v := data.(type) {
	case nil: // Ignore nil
	case string: // Ignore strings
	case float64:
		stat := Stat{
			Key:   key,
			Value: v,
		}
		ret = append(ret, stat)
	case []interface{}:
		for i, vv := range v {
			subKey := strconv.Itoa(i)
			if key != "" {
				subKey = fmt.Sprintf("%s.%s", key, subKey)
			}
			res := FlattenJSON(vv, subKey)
			ret = append(ret, res...)
		}
	case map[string]interface{}:
		for i, vv := range v {
			subKey := i
			if key != "" {
				subKey = fmt.Sprintf("%s.%s", key, subKey)
			}
			res := FlattenJSON(vv, subKey)
			ret = append(ret, res...)
		}
	}
	return ret
}
