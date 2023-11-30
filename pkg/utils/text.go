package utils

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Title returns a formatted string with the first letter in uppercase.
func Title(str string) string {
	return cases.Title(language.English, cases.NoLower).String(str)
}
