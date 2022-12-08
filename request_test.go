package wrpc

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestCreateBuilder(t *testing.T) {
	requestBase := RequestBase{
		Command: "",
		Id:      0,
		Args:    123,
	}
	_, _ = json.Marshal(requestBase)
	jsonStr := "{\"command\":\"\",\"id\":0,\"args\":\"1some-str\"}"

	var parsed RequestBase

	_ = json.Unmarshal([]byte(jsonStr), &parsed)

	var argsInt int = int(parsed.Args.(float64))

	fmt.Printf("%#v\n", argsInt)
}
