package main

import (
	"fmt"
	"os"
)

func main() {

	/*v, err := templating.ReadTemplate("templates/app.go")

	if err != nil {
		io.WriteString(os.Stdout, err.Error())
		return
	}

	fmt.Printf("%s", v.Render(&templating.TemplateData{
		Entries: map[string]any{
			"content": "inserted content",
			"entries": "",
			"entry":   "",
		},
	}))*/

	filename := os.Getenv("GOFILE")
	einfo, _ := ParseFile(filename)

	for _, info := range einfo {
		source := ParseOrCreateFile()
		fmt.Printf("Generating for %#v\n", info)
		result, err := info.RenderEverything(source)
		if err != nil {
			fmt.Printf("%sError :%v%s\n", SwitchToRed, err, ResetClr)
		}
		os.WriteFile("wrpc/app.go", []byte(result), 0644)
	}
}
