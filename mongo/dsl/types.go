package dsl

const (
	ConfigTypeString = "string"
	ConfigTypeNumber = "number"
	ConfigTypeBool   = "bool"
	ConfigTypeBytes  = "bytes"
)

type Config struct {
	TagName     string `json:"tag_name"`
	ColumnName  string `json:"column_name"`
	Type        string `json:"type"`
	Description string `json:"description"` // 描述
	Example     string `json:"example"`     // 用法
}
