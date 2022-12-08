package example

//go:generate go run ../generation/main.go
func Ping(request float64) int {
	return int(request + 1)
}
