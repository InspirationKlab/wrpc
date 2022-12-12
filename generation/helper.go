package main

import (
	"fmt"
	"strings"
)

func ExtractBetweenMarkers(source, marker string) ([]string, error) {
	markerInText := fmt.Sprintf("/*marker:%s*/", marker)
	first := 0
	first = strings.Index(source, markerInText)
	result := make([]string, 0, 10)
	for first != -1 {

		last := strings.Index(source[first+len(markerInText):], markerInText)
		if last == -1 {
			return nil, fmt.Errorf("unclosed marker %s", marker)
		}

		first = strings.Index(source[last+len(markerInText):], markerInText)
	}
	return result, nil
}

type Replace struct {
	content string
	begin   int
	end     int
}

func (r *Replace) Content() string {
	return r.content[r.begin:r.end]
}

func (r *Replace) WithReplacement(replacement string) string {
	return r.content[0:r.begin] + replacement + r.content[r.end:]
}

func ExtractSingle(source, marker string) (Replace, error) {
	markerInText := fmt.Sprintf("/*marker:%s*/", marker)
	first := strings.Index(source, markerInText)
	if first == -1 {
		return Replace{}, fmt.Errorf("marker %s not presented in source", marker)
	}
	last := strings.Index(source[first+len(markerInText):], markerInText) + first + len(markerInText)
	if last == -1 {
		return Replace{}, fmt.Errorf("unclosed marker %s", marker)
	}
	fmt.Printf("found begin=%d, end=%d\n", first, last)
	return Replace{
		content: source,
		begin:   first + len(markerInText),
		end:     last,
	}, nil
}
