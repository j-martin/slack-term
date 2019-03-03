package service

import (
	"testing"
)

func TestSanitizeLinks(t *testing.T) {
	assertFormat(t,
		"fff <https://github.com/alloytech/alloy/commit/0566e3fc4fba|1 new commit> pu",
		"fff < https://github.com/alloytech/alloy/commit/0566e3fc4fba |1 new commit> pu")
	assertFormat(t,
		"fff <https://github.com/alloytech/alloy/commit/0566e3fc4fba> pu",
		"fff < https://github.com/alloytech/alloy/commit/0566e3fc4fba > pu")
}

func assertFormat(t *testing.T, input string, expectedString string) {
	matchString := SanitizeLinks(input)
	if matchString != expectedString {
		t.Errorf("'%s' not equal to '%s'", matchString, expectedString)
	}
}
