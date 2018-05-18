// +build full

package lang

import (
	"tetra/lib/dbg"

	"golang.org/x/text/language"
)

func init() {
	matchlc = languageMatchLC
	validlc = languageValidLC
}

func languageMatchLC(availables, prefers []string) string {
	var p []language.Tag
	for _, s := range prefers {
		p = append(p, language.Make(s))
	}
	var a []language.Tag
	for _, s := range availables {
		x, err := language.Parse(s)
		if err != nil {
			dbg.Logf("failed to parse locale: %v\n", err)
			continue
		}
		a = append(a, x)
	}

	matcher := language.NewMatcher(a)
	tag, _, _ := matcher.Match(p...)
	return tag.String()
}

func languageValidLC(lc string) bool {
	_, err := language.Parse(lc)
	return err == nil
}
