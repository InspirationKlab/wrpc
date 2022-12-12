package example

import "bytes"

//go:generate ../generation/generation
func StreamMessages(request int, context, other AppContext, third bytes.Buffer) <-chan string {
	return make(chan string)
}

type AppContext struct {
	hitCount int
}
