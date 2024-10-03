package annotations

import (
	"bytes"
	"fmt"
	"strings"
	"text/scanner"
)

type status int

const (
	initial status = iota
	annotationName
	attributeName
	attributeValue
	done
)

func parseAnnotation(line string) (Annotation, error) {
	withoutComment := strings.TrimLeft(strings.TrimSpace(line), "/")

	firstKey := ""
	keys := bytes.Buffer{}
	fullLine := bytes.Buffer{}

	annotation := Annotation{
		Name:       "",
		LineValue:  "",
		Param:      "",
		Attributes: make(map[string]string),
	}

	var s scanner.Scanner
	s.Init(strings.NewReader(withoutComment))

	var tok rune
	var currentStatus status = initial
	var attrName string

	// czy szukamy reszty po @Anotacja ...
	// chodzi o to po spacji
	searchForRestLine := true
	isLineMode := false
	// onlyAnnName := ""
	for tok != scanner.EOF && currentStatus < done {
		tok = s.Scan()
		if isLineMode {
			fullLine.WriteString(s.TokenText())
			continue
		}
		switch tok {
		case '@':
			currentStatus = annotationName
		case '(':
			currentStatus = attributeName
		case '=':
			currentStatus = attributeValue
		case ',':
			// to dodaliśmy aby łatwo można było dodawać flagi
			// param1, param2, param3
			if currentStatus == attributeName {
				annotation.Attributes[strings.ToLower(attrName)] = ""
			}
			currentStatus = attributeName
		case ')':
			currentStatus = done
		case ':':
			//fmt.Printf( "XXX token=%v first=%v keys=`%v`, tok=`%s`\n", s.TokenText(), firstKey, keys.String(), tok)
			// @Anotacja: FULLLINE
			if searchForRestLine && currentStatus == annotationName {
				if len(line) > s.Column+3 {
					annotation.LineValue = strings.TrimSpace(line[s.Column+3:])
					annotation.IsFullLine = true
					return annotation, nil
				}
				/*
					 05.03.2018 zmienilismy - dziala ok
					 fmt.Printf("Name=%v cala ::: '%s' \n", annotation.Name, line, )
					currentStatus = attributeValue
					isLineMode=true
					searchForRestLine = false
				*/
			}
		case scanner.Ident:

			if firstKey == "" {
				firstKey = s.TokenText()
			} else {
				keys.WriteString(" ")
			}
			keys.WriteString(s.TokenText())
			// fmt.Printf("key:%s###%v\n", s.TokenText(),currentStatus)

			switch currentStatus {
			case annotationName:
				annotation.Name = s.TokenText()

			case attributeName:
				attrName = s.TokenText()
			}
		default:
			switch currentStatus {
			case attributeValue:
				if isLineMode {
					annotation.LineValue = s.TokenText() // strings.TrimSpace(s.TokenText())

				} else {
					annotation.Attributes[strings.ToLower(attrName)] = strings.Trim(s.TokenText(), "\"")
				}
			}
		}
	}

	if isLineMode && annotation.Name != "" {
		currentStatus = done
		annotation.IsFullLine = true
		annotation.LineValue = fullLine.String()

		//fmt.Printf("LINEMODE Name=%v keys=`%s`\n", annotation.Name, annotation.LineValue)
	}

	if currentStatus != done {

		// fmt.Printf("No i mamy coś nieskończonego %v, keys=%v aktualny statyus %v\n", annotation, keys.String(), currentStatus)
		if keys.String() == annotation.Name {
			annotation.IsFullLine = false
			annotation.LineValue = ""
			return annotation, nil
		}

		return annotation, fmt.Errorf("Invalid completion-status %v for annotation:%s",
			currentStatus, line)
	}
	return annotation, nil
}
