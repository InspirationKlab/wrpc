package example

//go:generate go run ../generation/main.go
func StreamMessages(request int, context AppContext) <-chan string {
	return make(chan string)
}

type AppContext struct {
	hitCount int
}
