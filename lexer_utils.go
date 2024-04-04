package main

import (
	"unicode"

	"golang.org/x/exp/slices"
)

func isLetterOrUnderscore(char rune) bool {
	return unicode.IsLetter(char) || char == '_'
}

func isAlphanumericOrUnderscore(char rune) bool {
	return unicode.IsLetter(char) || unicode.IsNumber(char) || char == '_'
}

func stringIsKeyword(token string) bool {
	keywords := []string{
		"select",
		"from",
		"where",
		"group",
		"by",
		"order",
		"asc",
		"desc",
		"limit",
		"offset",
		"create",
		"table",
		"insert",
		"into",
		"values",
		"text",
		"integer",
	}
	return slices.Contains(keywords, token)
}

func stringIsOperator(token string) bool {
	operators := []string{
		"=",
		"<>",
		">",
		">=",
		"<",
		"<=",
		"+",
		"-",
		"%",
	}
	return slices.Contains(operators, token)
}
