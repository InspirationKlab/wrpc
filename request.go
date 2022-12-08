package wrpc

type RequestBase struct {
	Command string `json:"command"`
	Id      int64  `json:"id"`
	Args    any    `json:"args"`
}

type Request[T any] struct {
	RequestBase
	Payload T `json:"payload"`
}
