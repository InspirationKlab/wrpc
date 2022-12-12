package wrpc

import (
	"encoding/json"
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

func AsTyped[T any](base *RequestBase) (T, error) {
	var t T
	err := json.Unmarshal([]byte(base.ArgStr), &t)
	if err != nil {
		return t, err
	}
	return t, nil
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
