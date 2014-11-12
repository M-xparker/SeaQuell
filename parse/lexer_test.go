package parse

import (
	"testing"
)

func Test_Run(t *testing.T) {
	l := lex("SELECT age FROM person;")
	items := []token{}
	for {
		i := l.nextItem()
		items = append(items, i)
		if i.typ == tokenEOF {
			break
		}
	}
	numberOfTokens := 9
	if len(items) != numberOfTokens {
		t.Errorf("Expected %d tokens; got %d", numberOfTokens, len(items))
	}
}
