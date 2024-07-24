package dsl

import (
	"errors"
	"regexp"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
)

type Exp interface {
	Name() string
}

type Rule struct {
	exps []Exp
}

type dslExp struct {
	p1 string
	p2 string
	p3 string
}

func (d dslExp) Name() string {
	return "dslExp"
}

type logicExp struct {
	p string
}

func (l logicExp) Name() string {
	return "logicExp"
}

type bracketExp struct {
	p string
}

func (b bracketExp) Name() string {
	return "bracketExp"
}

func TransFormExp(tokens []Token) (*Rule, error) {
	stream := newTokenStream(tokens)
	var ret []Exp
	for stream.hasNext() {
		tmpToken, err := stream.next()
		if err != nil {
			return nil, err
		}
		switch tmpToken.name {
		case tokenTag:
			p2, err := stream.next()
			if err != nil {
				return nil, err
			}
			if !(p2.name == tokenContains || p2.name == tokenFullEqual || p2.name == tokenNotEqual || p2.name == tokenRegexEqual) {
				return nil, errors.New("synax error in " + tmpToken.content + " " + p2.content)
			}
			p3, err := stream.next()
			if err != nil {
				return nil, err
			}
			if !(p3.name == tokenText || p3.name == tokenNumber) {
				return nil, errors.New("synax error in" + tmpToken.content + " " + p2.content + " " + p3.content)
			}
			// 正则缓存对象
			var dsl dslExp
			if p2.name == tokenRegexEqual {
				_, err := regexp.Compile(p3.content)
				if err != nil {
					return nil, err
				}
			}
			dsl = dslExp{p1: tmpToken.content, p2: p2.content, p3: p3.content}
			ret = append(ret, &dsl)
		case tokenAnd, tokenOr:
			exp := &logicExp{tmpToken.content}
			ret = append(ret, exp)
		case tokenLeftBracket, tokenRightBracket:
			exp := &bracketExp{tmpToken.content}
			ret = append(ret, exp)
		}
	}
	ret, err := infix2ToPostfix(ret)
	if err != nil {
		return nil, err
	}
	rule := new(Rule)
	rule.exps = ret
	return rule, nil
}

// 中缀表达式转换为后缀表达式
func infix2ToPostfix(exps []Exp) ([]Exp, error) {
	stack := NewStack()
	var ret []Exp
	for i := 0; i < len(exps); i++ {
		switch tmpExp := exps[i].(type) {
		case *dslExp:
			ret = append(ret, tmpExp)
		case *bracketExp:
			if tmpExp.p == tokenLeftBracket {
				// 左括号直接入栈
				stack.push(tmpExp)
			} else if tmpExp.p == tokenRightBracket {
				// 右括号则弹出元素,直到遇到左括号
				for !stack.isEmpty() {
					pre, exist := stack.top().(*bracketExp)
					if exist && pre.p == tokenLeftBracket {
						stack.pop()
						break
					}
					ret = append(ret, stack.pop().(Exp))
				}
			}
		case *logicExp:
			if !stack.isEmpty() {
				top := stack.top()
				bracket, exist := top.(*bracketExp)
				if exist && bracket.p == tokenLeftBracket {
					stack.push(tmpExp)
					continue
				}
				ret = append(ret, top.(Exp))
				stack.pop()
			}
			stack.push(tmpExp)
		default:
			return nil, errors.New("unknown transform type")
		}
	}
	for !stack.isEmpty() {
		tmp := stack.pop()
		ret = append(ret, tmp.(Exp))
	}
	return ret, nil
}

func handleType(text string, currentConfig *Config) (interface{}, error) {
	var v interface{}
	switch currentConfig.Type {
	case ConfigTypeString:
		v = text
	case ConfigTypeNumber:
		v, _ = strconv.Atoi(text)
	case ConfigTypeBool:
		if text == "true" || text == "TRUE" {
			v = true
		} else {
			v = false
		}
	}
	return v, nil
}

func (r *Rule) ToMongo(configs []Config) (bson.D, error) {
	stack := NewStack()
	for i := 0; i < len(r.exps); i++ {
		switch next := r.exps[i].(type) {
		case *dslExp:
			s1 := next.p1 // tag name
			var currentConfig *Config

			for _, config := range configs {
				if config.TagName == s1 {
					currentConfig = &config
					break
				}
			}

			if currentConfig == nil {
				return nil, errors.New("unknown tag name")
			}

			var r bson.D
			text := next.p3 // value
			if currentConfig.ColumnName != "" {
				s1 = currentConfig.ColumnName // 设定column
			}

			switch next.p2 {
			case tokenFullEqual:
				v, err := handleType(text, currentConfig)
				if err != nil {
					return nil, err
				}
				r = bson.D{
					{
						Key: s1,
						Value: bson.M{
							"$eq": v,
						},
					},
				}
			case tokenContains:
				if currentConfig.Type == ConfigTypeString {
					text = regexp.QuoteMeta(text)
				}
				v, err := handleType(text, currentConfig)
				if err != nil {
					return nil, err
				}
				if currentConfig.Type == ConfigTypeString {
					r = bson.D{
						{
							Key:   s1,
							Value: bson.M{"$regex": v, "$options": "i"},
						},
					}
				} else {
					r = bson.D{
						{
							Key:   s1,
							Value: v,
						},
					}
				}
			case tokenNotEqual:
				if currentConfig.Type == ConfigTypeString {
					text = regexp.QuoteMeta(text)
				}
				v, err := handleType(text, currentConfig)
				if err != nil {
					return nil, err
				}
				if currentConfig.Type == ConfigTypeString {
					r = bson.D{
						{
							Key: s1,
							Value: bson.M{"$not": bson.M{
								"$regex":   v,
								"$options": "i",
							}},
						},
					}
				} else {
					r = bson.D{
						{
							Key:   s1,
							Value: bson.M{"$not": v},
						},
					}
				}
			case tokenRegexEqual:
				if currentConfig.Type == ConfigTypeString {
					text = regexp.QuoteMeta(text)
				} else {
					return nil, errors.New("regex type not support ~=")
				}
				// 小写模糊
				r = bson.D{
					{
						Key:   s1,
						Value: bson.M{"$regex": text, "$options": "i"},
					},
				}
			default:
				panic("unknown p2 token")
			}
			stack.push(r)
		case *logicExp:
			p1 := stack.pop().(bson.D)
			p2 := stack.pop().(bson.D)
			var r bson.D
			switch next.p {
			case tokenAnd:
				r = bson.D{
					{
						Key:   "$and",
						Value: []bson.D{p1, p2},
					},
				}
			case tokenOr:
				r = bson.D{
					{
						Key:   "$or",
						Value: []bson.D{p1, p2},
					},
				}
			default:
				panic("unknown logic type")
			}
			stack.push(r)
		default:
			panic("error eval")
		}
	}
	top := stack.top().(bson.D)
	return top, nil
}

func Dsl2Mongo(dsl string, configs []Config) (bson.D, error) {
	tokens, err := ParseTokens(dsl)
	if err != nil {
		return nil, err
	}
	exp, err := TransFormExp(tokens)
	if err != nil {
		return nil, err
	}
	return exp.ToMongo(configs)
}
