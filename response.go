package wrpc

type Response struct {
	Id   int64 `json:"id"`
	Data any   `json:"data"`
}
