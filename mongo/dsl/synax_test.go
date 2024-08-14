package dsl

import (
	"testing"

	"github.com/kr/pretty"
)

var webSearchConfigs = []Config{
	{
		TagName:     "domain",
		ColumnName:  "domain",
		Description: "通过域名查询",
		Example:     "domain=\"example.com\"",
		Type:        ConfigTypeString,
	},
	{
		TagName:     "favicon_md5",
		ColumnName:  "faviconmd5",
		Description: "通过favicon_md5查询",
		Example:     "favicon_md5=\"f0d8af658c4f0c0d82b3c7d658be53d8\"",
		Type:        ConfigTypeString,
	},
	{
		TagName:     "favicon_mmh3",
		ColumnName:  "faviconmmh3",
		Description: "通过favicon_mmh3查询",
		Example:     "favicon_mmh3=\"1300889937\"",
		Type:        ConfigTypeString,
	},
	{
		TagName:     "header",
		ColumnName:  "headers",
		Description: "查询header头",
		Example:     "header=\"x-via\"",
		Type:        ConfigTypeString,
	},
	{
		TagName:     "ip",
		ColumnName:  "ip",
		Description: "通过ip查询",
		Example:     "ip=\"1.1.1.1\"",
		Type:        ConfigTypeString,
	},
	{
		TagName:     "country",
		ColumnName:  "host.country",
		Description: "通过country查询",
		Example:     "country=\"中国\"",
		Type:        ConfigTypeString,
	},
	{
		TagName:     "port",
		ColumnName:  "port",
		Description: "通过port查询",
		Example:     "port=\"80\"",
		Type:        ConfigTypeString,
	},
	{
		TagName:     "resp",
		ColumnName:  "resp",
		Description: "查询返回包",
		Example:     "resp=\"<html>\"",
		Type:        ConfigTypeString,
	},
	{
		TagName:     "scheme",
		ColumnName:  "scheme",
		Description: "查询http协议名称",
		Example:     "scheme=\"http\"",
		Type:        ConfigTypeString,
	},
	{
		TagName:     "server",
		ColumnName:  "server",
		Description: "server查询",
		Example:     "server=\"nginx\"",
		Type:        ConfigTypeString,
	},
	{
		TagName:     "status_code",
		ColumnName:  "statuscode",
		Description: "状态码查询",
		Example:     "status_code=200",
		Type:        ConfigTypeNumber,
	},
	{
		TagName:     "app",
		ColumnName:  "tags.content",
		Description: "查询指纹",
		Example:     "app=\"nginx\"",
		Type:        ConfigTypeString,
	},
	{
		TagName:     "tag",
		ColumnName:  "tags.name",
		Description: "通过tag查询",
		Example:     "tag=\"名称\"",
		Type:        ConfigTypeString,
	},
	{
		TagName:     "title",
		ColumnName:  "title",
		Description: "查询标题",
		Example:     "title=\"nginx\"",
	},
	{
		TagName:     "url",
		ColumnName:  "url",
		Description: "通过url查询",
		Example:     "url=\"example.com\"",
		Type:        ConfigTypeString,
	},
	{
		TagName:     "is_req",
		ColumnName:  "is_req",
		Description: "通过类型查询",
		Example:     "url=true",
		Type:        ConfigTypeBool,
	},
}

func TestTransFormExp(t *testing.T) {
	s := "resp=\"powered by\" && is_req!=true && app=\"Meta-Author\""
	tokens, err := ParseTokens(s)
	if err != nil {
		t.Fatal(err)
	}
	exp, err := TransFormExp(tokens)
	if err != nil {
		t.Fatal(err)
	}
	dsl, err := exp.ToMongo(webSearchConfigs)
	if err != nil {
		t.Fatal(err)
	}
	pretty.Println(dsl)
}
