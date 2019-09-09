package stats

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

func TestFlattenJSON(t *testing.T) {
	tables := []struct {
		input string
		ret   []Stat
	}{
		{
			"{ \"a\": 1, \"b\": { \"a\": 43.44, \"b\": -123.25, \"c\": [1,2,3, { \"a\": 1, \"b\": 1.25 }] }, \"c\": \"def\" }",
			[]Stat{
				Stat{"a", 1, nil},
				Stat{"b.a", 43.44, nil},
				Stat{"b.b", -123.25, nil},
				Stat{"b.c.0", 1, nil},
				Stat{"b.c.1", 2, nil},
				Stat{"b.c.2", 3, nil},
				Stat{"b.c.3.a", 1, nil},
				Stat{"b.c.3.b", 1.25, nil},
			},
		},
	}

	for _, table := range tables {
		var data interface{}
		err := json.Unmarshal([]byte(table.input), &data)
		if err != nil {
			t.Errorf("Parse Failed")
		}
		ret := FlattenJSON(data, "")
		sort.Slice(ret, func(i, j int) bool { return ret[i].Key < ret[j].Key })
		if !reflect.DeepEqual(ret, table.ret) {
			t.Errorf("FlattenJSON(%s) was incorrect, \ngot: %v, \nwant: %v", table.input, ret, table.ret)
		}
	}
}
