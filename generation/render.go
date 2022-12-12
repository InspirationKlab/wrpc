package main

import (
	"fmt"
	"os"
	"strings"
)

type DependencyTypename = string

type RenderContext struct {
	// typename -> name
	Deps    map[DependencyTypename]string
	Builder strings.Builder
}

func (info *ExtendedFunctionInfo) BuildDependencies(context *RenderContext) {
	for _, dependency := range info.Dependencies {
		if _, presented := context.Deps[dependency.Render()]; !presented {
			context.Deps[dependency.Render()] = dependency.Name
		}
	}
}

func (r *RenderContext) RenderDependencies() string {
	entries := make([]string, 0, len(r.Deps))
	for typename, name := range r.Deps {
		entries = append(entries, fmt.Sprintf("%s %s", name, typename))
	}
	return strings.Join(entries, ", ")
}

func (info *ExtendedFunctionInfo) RenderInvocation(context *RenderContext) {

	argsRender := ""
	argsRender += "typedRequestArg"

	for _, dependency := range info.Dependencies {
		argsRender += fmt.Sprintf(", %s", context.Deps[dependency.Render()])
	}

	if !info.IsVoid {
		const template = `result := %s.%s(%s)`
		context.Builder.WriteString(fmt.Sprintf(template, info.Package, info.Name, argsRender))
	} else {
		const template = `%s.%s(%s)`
		context.Builder.WriteString(fmt.Sprintf(template, info.Package, info.Name, argsRender))
	}
}

func (info *ExtendedFunctionInfo) RenderSwitchCase(context *RenderContext) {

	const templatePreamble = `
				case "%s":
					go func (request wrpc.RequestBase) {
						typedRequestArg, err := wrpc.AsTyped[%s](&request)
						if err != nil {
							outChannel <- wrpc.Response {
								Id: request.Id,
								Error: "Error in parsing json: " + err.Error(),
							}
							return
						}
`
	const voidTemplate = `
						outChannel <- wrpc.Response {
							Id: request.Id,
						}
`
	const syncTemplate = `
						outChannel <- wrpc.Response{
							Id: request.Id,
							Data: result,
						}
`
	const asyncTemplate = `
						for rEntry := range result {
							outChannel <- wrpc.Response{
								Id: request.Id,
								Data: rEntry,
							}
						}
`
	const postAmble = `
					}(parsed)
`
	context.Builder.WriteString(fmt.Sprintf(templatePreamble, info.Name, info.RequestParameter.Render()))
	info.RenderInvocation(context)

	if info.IsVoid {
		context.Builder.WriteString(voidTemplate)
	} else {
		if info.IsAsync {
			context.Builder.WriteString(asyncTemplate)
		} else {
			context.Builder.WriteString(syncTemplate)
		}
	}
	context.Builder.WriteString(postAmble)
}

func parseDependencies(content string) map[DependencyTypename]string {
	entries := strings.Split(content, ",")
	result := make(map[DependencyTypename]string)
	for _, entry := range entries {
		nameAndType := strings.Split(strings.TrimSpace(entry), " ")

		if len(nameAndType) != 2 {
			continue
		}

		name := nameAndType[0]
		_type := nameAndType[1]
		result[_type] = name
	}
	return result
}

const bodyTemplate = `
package wrpc_app

import (
	"encoding/json"
	"github.com/InspirationKlab/wrpc"
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)

func WsAppEntry(/*marker:deps*//*marker:deps*/) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request){
		upgrade := websocket.Upgrader {CheckOrigin: func(r *http.Request) bool {
			return true
		}}
		connection, err := upgrade.Upgrade(writer, request, nil)
		if err != nil {
			http.Error(writer, err.Error(), 500)
		}
		outChannel := make(chan wrpc.Response)
		defer connection.Close()

		go func() {
			for err == nil {
				mt, message, err := connection.ReadMessage()
				if err != nil || mt == websocket.CloseMessage {
					return
				}
				var parsed wrpc.RequestBase			
				json.Unmarshal(message, &parsed)
				switch parsed.Command {
					/*marker:switch*/
					/*marker:switch*/
				}
			}

		}()
		for {
			select {
			case output := <-outChannel:
				err = connection.WriteJSON(output)
				if err != nil {
					log.Printf("Error in writing message to websocket, %v", err)
					return
				}
			}
		}
	}
}

`

func ParseOrCreateFile() string {
	_, err := os.Stat("wrpc")
	if err != nil {
		err := os.Mkdir("wrpc", 0777)
		if err != nil {
			fmt.Printf("%sError in creating dir%v%s\n", SwitchToRed, err, ResetClr)
		}
	}
	_, err = os.Stat("wrpc/app.go")
	if err != nil {
		_ = os.WriteFile("wrpc/app.go", []byte(bodyTemplate), 0644)
	}

	bytes, _ := os.ReadFile("wrpc/app.go")
	return string(bytes)
}

func (info *ExtendedFunctionInfo) RenderEverything(source string) (string, error) {
	depString, err := ExtractSingle(source, "deps")
	if err != nil {
		return "", err
	}
	deps := parseDependencies(depString.Content())

	rContext := RenderContext{Deps: deps}
	info.BuildDependencies(&rContext)

	source = depString.WithReplacement(rContext.RenderDependencies())
	swicthRepl, err := ExtractSingle(source, "switch")
	if err != nil {
		return "", err
	}
	info.RenderSwitchCase(&rContext)
	return swicthRepl.WithReplacement(swicthRepl.Content() + "\n" + rContext.Builder.String()), nil
}
