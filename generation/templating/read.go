package templating

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	parseModeText   = iota
	parseModeMarker = iota
)

func padding(count int) string {
	s := ""
	for i := 0; i < count; i++ {
		s += " "
	}
	return s
}

func printErrorMessage(content string, lineno, linepos int, err string) string {
	lines := strings.Split(content, "\n")
	errorLine := lines[lineno]
	return fmt.Sprintf("%s\n%s^\n%s%s\n", errorLine, padding(linepos), padding(linepos), err)
}

func ReadTemplate(filename string) (ParsedTemplate, error) {
	nodes := make([]templateNode, 0, 100)
	contentBytes, err := os.ReadFile(filename)
	if err != nil {
		return ParsedTemplate{}, err
	}
	content := string(contentBytes)

	parseMode := parseModeText

	lineno := 0
	linepos := 0

	var currentNode = &templateText{}

	for i := 0; i < len(content); i++ {
		strView := content[i:]

		if content[i] == '\n' {
			lineno++
			linepos = 0
		}

		if content[i] == '\t' {
			linepos += 2
		}

		linepos++

		switch parseMode {
		case parseModeText:
			if strings.HasPrefix(strView, "/*marker:") {

				nodes = append(nodes, currentNode)

				currentNode = &templateText{}

				end := strings.Index(strView, "*/") - len("/*marker:")
				linepos += len("/*marker:")
				if end < 0 {
					return ParsedTemplate{}, errors.New(printErrorMessage(content, lineno, linepos, "unterminated marker"))
				}

				tokenStart := i + len("/*marker:")
				position := 0
				for position < end && content[tokenStart+position] != ':' {
					position++
				}
				token := content[tokenStart : tokenStart+position]
				if token == "" {
					return ParsedTemplate{}, errors.New(printErrorMessage(content, lineno, linepos, "marker without content"))
				}

				//TODO: read qualifiers

				nodes = append(nodes, &templateMarkerInsert{name: token})

				linepos += end
				i += len("/*marker:") + end + 1
			} else {
				currentNode.text += string(content[i])
			}
		}
	}

	nodes = append(nodes, currentNode)

	return ParsedTemplate{nodes: nodes}, nil
}
