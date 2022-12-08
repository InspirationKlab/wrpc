package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type FuncInfo struct {
	Name         string
	ParamType    string
	ReturnType   string
	Dependencies map[string]string
}

const switchCaseTemplate = `
		case "%s":
			%s(args)
			break
`

const switchCaseSyncTemplate = `
		case "%s":
				go func(request wrpc.RequestBase) {
					typedArg := wrpc.AsTyped[%s](&request)
					result := %s.%s(%s)
					outChannel <- wrpc.Response{
						Id: request.Id,
						Data: result,
					}
				}(parsed)
			break
`

const switchCaseAsyncTemplate = `
			case "%s":
				go func(request wrpc.RequestBase) {
					typedArg := wrpc.AsTyped[%s](&request)
					result := %s.%s(%s)
					for {
						select {
						case value:= <- result:
							outChannel <- wrpc.Response{
								Id: request.Id,
								Data: value,
							}
						}
					}
				}(parsed)
				break
`

const template = `
package wrpc_app

import "blacksec.com/wrpc/v2"

func WsAppEntry(/* wrpc-app-deps:begin */   /* wrpc-app-deps:end */) func(writer http.ResponseWriter, request *http.Request) {
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
					return // Выходим из цикла, если клиент пытается закрыть соединение или связь прервана
				}
				var parsed wrpc.RequestBase			
				json.Unmarshal(message, &parsed)
				switch parsed.Command {
					// wrpc-app-list
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

func processTypeCoercion(argType string) (jsonArgType string) {

	numericTypes := map[string]bool{
		"int":     true,
		"int64":   true,
		"float32": true,
		"int32":   true,
	}

	jsonArgType = argType

	if numericTypes[argType] {
		jsonArgType = "float64"
	}

	return jsonArgType
}

func toCamelCase(source string) string {
	words := strings.Split(source, " ")
	key := strings.ToLower(words[0])
	for _, word := range words[1:] {
		key += strings.Title(word)
	}
	return key
}

func renderInvocation(info FuncInfo) string {
	params := []string{"typedArg"}

	for name := range info.Dependencies {
		params = append(params, name)
	}

	return strings.Join(params, ", ")
}

func functionInfoToText(info FuncInfo) string {

	if strings.Contains(info.ReturnType, "chan") {
		return fmt.Sprintf(switchCaseAsyncTemplate, info.Name, info.ParamType, os.Getenv("GOPACKAGE"), info.Name, renderInvocation(info))
	}

	return fmt.Sprintf(switchCaseSyncTemplate, info.Name, info.ParamType, os.Getenv("GOPACKAGE"), info.Name, renderInvocation(info))

}

var builtInTypes = map[string]bool{
	"int":     true,
	"int32":   true,
	"float32": true,
	"int64":   true,
	"float64": true,
	"string":  true,
	"[]byte":  true,
}

func getCurrentDeps(fileContent string) map[string]string {
	begin := strings.Index(fileContent, "/* wrpc-app-deps:begin */")

	if begin == -1 {
		log.Fatalln("could not detected dependency place in file!")
	}
	begin += len("/* wrpc-app-deps:begin */") + 1

	end := strings.Index(fileContent, "/* wrpc-app-deps:end */")
	if end == -1 {
		log.Fatalln("could not detected dependency place in file!")
	}
	depStr := strings.TrimSpace(fileContent[begin:end])
	log.Printf("Dep str is \"%s\"\n", depStr)
	deps := strings.Split(depStr, ",")
	result := make(map[string]string)
	for _, dep := range deps {
		dep = strings.TrimSpace(dep)
		if dep == "" {
			continue
		}

		name, _type := toNameAndType(dep)
		result[name] = _type
	}
	log.Printf("Deps are %#v\n", deps)
	return result
}

func depsToStr(deps map[string]string) string {
	result := ""

	for name, _type := range deps {
		if result != "" {
			result += ", "
		}
		result += name + " " + _type
	}

	return result
}

func insertDeps(content string, deps map[string]string) string {
	begin := strings.Index(content, "/* wrpc-app-deps:begin */")

	if begin == -1 {
		log.Fatalln("could not detected dependency place in file!")
	}
	begin += len("/* wrpc-app-deps:begin */") + 1

	end := strings.Index(content, "/* wrpc-app-deps:end */")
	if end == -1 {
		log.Fatalln("could not detected dependency place in file!")
	}
	return content[0:begin] + depsToStr(deps) + content[end:]
}

func toNameAndType(param string) (string, string) {
	paramNameAndType := strings.Split(strings.TrimSpace(param), " ")
	return paramNameAndType[0], strings.Join(paramNameAndType[1:], " ")
}

func parseFuncInfo(line string) FuncInfo {
	regex, err := regexp.Compile("func (.*)\\((.*)\\) (.*) *{$")
	if err != nil {
		log.Fatalln(err)
	}
	groups := regex.FindStringSubmatch(line)

	args := strings.Split(groups[2], ",")

	paramNameAndType := strings.Split(args[0], " ")
	deps := make(map[string]string)

	for i := 1; i < len(args); i++ {
		name, _type := toNameAndType(args[i])

		deps[name] = _type
	}

	info := FuncInfo{
		Name:         groups[1],
		ParamType:    paramNameAndType[1],
		ReturnType:   groups[3],
		Dependencies: deps,
	}
	fmt.Printf("%#v", info)
	return info
}

func main() {
	filename := os.Getenv("GOFILE")
	fmt.Printf("Running %s go on %s line = %s\n", os.Args[0], os.Getenv("GOFILE"), os.Getenv("GOLINE"))
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalln("Cannot open file", filename)
	}
	lineno, _ := strconv.Atoi(os.Getenv("GOLINE"))
	bytes, _ := ioutil.ReadAll(file)

	allLines := strings.Split(string(bytes), "\n")
	info := parseFuncInfo(allLines[lineno])

	_ = file.Close()
	_, err = os.Stat("wrpc")

	if err != nil {
		fmt.Println("Creating file:", os.Mkdir("wrpc", 0777))
	}

	integrated := functionInfoToText(info)

	_, err = os.Stat("wrpc/app.go")
	if err != nil {
		_ = os.WriteFile("wrpc/app.go", []byte(template), 0644)
	}

	contents, err := os.ReadFile("wrpc/app.go")

	if err != nil {
		log.Fatalln("cannot access root file")
	}

	lines := strings.Split(string(contents), "\n")
	resultingLines := make([]string, 0, 100)
	for _, line := range lines {
		resultingLines = append(resultingLines, line)
		if strings.Contains(line, "// wrpc-app-list") {
			resultingLines = append(resultingLines, integrated)
			fmt.Printf("%s\n", integrated)
		}
	}
	content := strings.Join(resultingLines, "\n")
	currentDeps := getCurrentDeps(content)

	for name, _type := range info.Dependencies {
		currentDeps[name] = _type
	}

	content = insertDeps(content, currentDeps)

	os.WriteFile("wrpc/app.go", []byte(content), 0644)
}
