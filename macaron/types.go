package macaron

import (
	"encoding/json"
	"fmt"
)

type Result map[string]interface{}

var OK = Result{"result": "ok"}

func (r Result) BeforeWrite() interface{} {
	if _, ok := r["result"]; !ok {
		r["result"] = "ok"
	}
	return r
}

type Error struct {
	Result  string `json:"result"`
	Message string `json:"msg"`
	Code    int    `json:"-"`
}

func NewError(result, msg string, code int) *Error {
	return &Error{result, msg, code}
}

func (e *Error) GetStatusCode() int {
	return e.Code
}

func (e *Error) Error() string {
	return e.JsonString()
}

func (e *Error) JsonString() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func (e *Error) Msg(msg string) *Error {
	var err = *e
	err.Message = msg
	return &err
}

type FileResp struct {
	Name        string
	ContentType string
	Content     []byte
}

func NewImageResp(ext string, content []byte) *FileResp {
	return &FileResp{
		ContentType: fmt.Sprintf("image/%s", ext),
		Content:     content,
	}
}

func NewFileResp(name string, content []byte) *FileResp {
	return &FileResp{
		Name:        name,
		ContentType: "application/octet-stream",
		Content:     content,
	}
}

func (f *FileResp) Get() (name, contentType string, content []byte) {
	return f.Name, f.ContentType, f.Content
}
