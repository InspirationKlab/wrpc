package wrpc

import (
	"encoding/json"
	"log"
)

type RequestBase struct {
	Command string `json:"command"`
	Id      int64  `json:"id"`
	ArgStr  string
}

type Request[T any] struct {
	RequestBase
	Payload T `json:"payload"`
}

func AsTyped[T any](base *RequestBase) T {
	var t T
	err := json.Unmarshal([]byte(base.ArgStr), &t)
	if err != nil {
		log.Printf("Error deserializing %v\n", err)
	}
	return t
}

func (base *RequestBase) UnmarshalJSON(source []byte) error {
	var dat map[string]*json.RawMessage

	if err := json.Unmarshal(source, &dat); err != nil {
		return err
	}

	if err := json.Unmarshal(*dat["command"], &base.Command); err != nil {
		return err
	}

	if err := json.Unmarshal(*dat["id"], &base.Id); err != nil {
		return err
	}

	base.ArgStr = string(*dat["args"])

	return nil
}
