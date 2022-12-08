package wrpc

import (
	"log"
	"reflect"
)

type appBuilder struct {
	handlers map[string]func(any) any
}

type ServerBuilder interface {
	Map(command string, process func(any) any)
}

func CreateBuilder() ServerBuilder {
	return &appBuilder{
		handlers: make(map[string]func(any) any),
	}
}

func (builder *appBuilder) Map(command string, process func(any) any) {
	log.Println(reflect.TypeOf(process))
}
