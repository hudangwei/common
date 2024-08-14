package dsl

import (
	"errors"
	"regexp"
	"strings"
)

//	title="NBX NetSet" || (header="Alternates" && body="NBX")
//	1||(2&&3) => 1 2 3 && ||
//	header="X-Copyright: wspx" || header="X-Powered-By: ANSI C"
//	header="SS_MID" && header="squarespace.net"
//	expr = self.in2post(expr)
//	print("后缀表达式", expr)

// 支持 title body header icon
// 符号支持 && || ()
type Token struct {
	name    string
	content string
}

const (
	tokenTag    = "tag"
	tokenText   = "text"
	tokenNumber = "number"
	tokenBool   = "bool"

	tokenContains   = "="
	tokenFullEqual  = "=="
	tokenNotEqual   = "!="
	tokenRegexEqual = "~="

	tokenAnd = "&&"
	tokenOr  = "||"

	tokenLeftBracket  = "("
	tokenRightBracket = ")"
)

func ParseTokens(s1 string) ([]Token, error) {
	var s []rune = []rune(s1)
	var tokens []Token
	i := 0
	var tmpToken Token
	pattern := `^\w[\d\w_]*`
	defaultReg, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	for i < len(s) {
		x := s[i]
		switch x {
		case '"':
			n := []rune{}
			i2 := i + 1
			for i2 < len(s) {
				if s[i2] == '\\' {
					n = append(n, s[i2+1])
					i2 += 1
				} else if s[i2] == '"' {
					tmpToken = Token{
						name:    tokenText,
						content: string(n),
					}
					tokens = append(tokens, tmpToken)
					i = i2
					goto end
				} else {
					n = append(n, s[i2])
				}
				i2 += 1
			}
			return nil, errors.New("unknown text:" + string(s[i:]))
		end:
			i += 1
		case '=':
			if string(s[i:i+2]) == "==" {
				tmpToken = Token{
					name:    tokenFullEqual,
					content: "==",
				}
				i += 1
			} else {
				tmpToken = Token{
					name:    tokenContains,
					content: "=",
				}
			}
			tokens = append(tokens, tmpToken)
			i += 1
		case '~':
			if string(s[i:i+2]) == "~=" {
				tmpToken = Token{
					name:    tokenRegexEqual,
					content: "~=",
				}
				tokens = append(tokens, tmpToken)
				i += 2
			}
		case '!':
			if string(s[i:i+2]) == "!=" {
				tmpToken = Token{
					name:    tokenNotEqual,
					content: "!=",
				}
				tokens = append(tokens, tmpToken)
				i += 2
			} else {
				return nil, errors.New("unknown text:" + string(s[i:]))
			}
		case '|':
			if string(s[i:i+2]) == "||" {
				tmpToken = Token{
					name:    tokenOr,
					content: "||",
				}
				tokens = append(tokens, tmpToken)
			}
			i += 2
		case '&':
			if string(s[i:i+2]) == "&&" {
				tmpToken = Token{
					name:    tokenAnd,
					content: "&&",
				}
				tokens = append(tokens, tmpToken)
			}
			i += 2
		case '(':
			tmpToken = Token{
				name:    tokenLeftBracket,
				content: "(",
			}
			tokens = append(tokens, tmpToken)
			i += 1
		case ')':
			tmpToken = Token{
				name:    tokenRightBracket,
				content: ")",
			}
			tokens = append(tokens, tmpToken)
			i += 1
		case ' ':
			i += 1
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
			tmp := []rune{s[i]}
			i2 := i + 1
			for i2 < len(s) {
				if (s[i2] >= '0' && s[i2] <= '9') || s[i2] == '-' {
					tmp = append(tmp, s[i2])
					i2 += 1
					i += 1
				} else {
					break
				}
			}
			tmpToken = Token{
				name:    tokenNumber,
				content: string(tmp),
			}
			tokens = append(tokens, tmpToken)
			i += 1
		default:
			textOption := string(s[i:])
			ret := defaultReg.FindString(textOption)
			if ret != "" {
				lowerRet := strings.ToLower(ret)
				if lowerRet == "true" || lowerRet == "false" {
					tmpToken = Token{
						name:    tokenBool,
						content: lowerRet,
					}
				} else {
					tmpToken = Token{
						name:    tokenTag,
						content: ret,
					}
				}
				tokens = append(tokens, tmpToken)
				i += len(ret)
			} else {
				return nil, errors.New("unknown text:" + string(s[i:]))
			}
		}
	}
	return tokens, nil
}
