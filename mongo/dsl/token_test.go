package dsl

import (
	"testing"

	"github.com/kr/pretty"
)

func TestParseTokens1(t *testing.T) {
	tokens := []string{
		"header=\"realmComtrend Gigabit 802.11n Router\" || banner=\"Comtrend Gigabit 802.11n Router\" && z=1",
		"(title=\"你好\" && statuscode=200)||tag=\"x\"",
	}
	for _, token := range tokens {
		ss, err := ParseTokens(token)
		if err != nil {
			t.Fatal(err)
		}
		pretty.Println(ss)
	}
}

func TestParseTokens2(t *testing.T) {
	s := "body~=\"(<center><strong>EZCMS ([\\d\\.]+) )\""
	tokens, err := ParseTokens(s)
	if err != nil {
		t.Fatal(err)
	}
	pretty.Println(tokens)
}
