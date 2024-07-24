package dsl

import "errors"

type tokenStream struct {
	tokens      []Token
	index       int
	tokenLength int
}

func newTokenStream(tokens []Token) *tokenStream {
	ret := new(tokenStream)
	ret.tokens = tokens
	ret.tokenLength = len(tokens)
	return ret
}

func (this *tokenStream) rewind() {
	this.index -= 1
}

func (this *tokenStream) next() (Token, error) {
	if this.index >= len(this.tokens) {
		return Token{}, errors.New("token解析失败")
	}
	token := this.tokens[this.index]
	this.index += 1
	return token, nil
}

func (this tokenStream) hasNext() bool {
	return this.index < this.tokenLength
}
