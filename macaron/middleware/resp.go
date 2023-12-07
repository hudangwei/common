package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/hudangwei/common/macaron"
)

type RawHttpResponse interface {
	WriteRawResponse(http.ResponseWriter) error
}

type HttpFileResponse interface {
	Get() (name, contentType string, content []byte)
}

type SpecCodeHttpResponse interface {
	GetStatusCode() int
}

type SpecContentTypeHttpResponse interface {
	GetContentType() (contentType string, content []byte)
}

type WriteHttpResponse interface {
	BeforeWrite() interface{}
}

func HTTPResp() macaron.ReturnHandler {
	return func(ctx *macaron.Context, vals []reflect.Value) {
		//处理handler函数的返回值
		if len(vals) > 0 {
			ret := vals[0].Interface()
			// pretty.Println(ctx.Req.URL.Path, ret)
			switch v := ret.(type) {
			case nil:
				return
			case RawHttpResponse:
				rr := ret.(RawHttpResponse)
				if err := rr.WriteRawResponse(ctx.RespWriter); err != nil {
					logWriteErrors(ctx.Req, err)
				}
			case HttpFileResponse:
				fr := ret.(HttpFileResponse)
				name, contentType, content := fr.Get()
				if contentType == "" {
					contentType = "application/octet-stream"
				}
				ctx.RespWriter.Header().Set("Content-type", contentType)
				if len(name) > 0 {
					ctx.RespWriter.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", name))
				}
				_, err := ctx.RespWriter.Write(content)
				if err != nil {
					logWriteErrors(ctx.Req, err)
				}
			case SpecCodeHttpResponse:
				resp := ret.(SpecCodeHttpResponse)
				if err := writeJsonToResp(ctx.RespWriter, resp.GetStatusCode(), resp); err != nil {
					logWriteErrors(ctx.Req, err)
				}
			case SpecContentTypeHttpResponse:
				resp := ret.(SpecContentTypeHttpResponse)
				contentType, body := resp.GetContentType()
				ctx.RespWriter.Header().Set("Content-Type", contentType)

				ctx.RespWriter.WriteHeader(200)
				if _, err := ctx.RespWriter.Write(body); err != nil {
					logWriteErrors(ctx.Req, err)
				}
			case WriteHttpResponse:
				resp := ret.(WriteHttpResponse).BeforeWrite()
				if err := writeJsonToResp(ctx.RespWriter, http.StatusOK, resp); err != nil {
					logWriteErrors(ctx.Req, err)
				}
			case map[string]interface{}:
				if _, ok := v["result"]; !ok {
					v["result"] = "ok"
				}
				if err := writeJsonToResp(ctx.RespWriter, http.StatusOK, v); err != nil {
					logWriteErrors(ctx.Req, err)
				}
			case error:
				if err := writeJsonToResp(ctx.RespWriter, http.StatusInternalServerError, macaron.NewError("error", v.Error(), http.StatusInternalServerError)); err != nil {
					logWriteErrors(ctx.Req, err)
				}
			case []byte:
				ctx.RespWriter.Header().Set("Content-Type", "application/json")

				ctx.RespWriter.WriteHeader(200)
				if _, err := ctx.RespWriter.Write(ret.([]byte)); err != nil {
					logWriteErrors(ctx.Req, err)
				}
			default:
				if err := writeJsonToResp(ctx.RespWriter, http.StatusOK, ret); err != nil {
					logWriteErrors(ctx.Req, err)
				}
			}
		}
	}
}

func logWriteErrors(req *http.Request, err error) {
	log.Println("write resp failed",
		"err", err,
		"url", req.URL.String())
}

func writeJsonToResp(rw http.ResponseWriter, code int, data interface{}) error {
	bs, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}
	rw.Header().Set("Content-Type", "application/json")

	rw.WriteHeader(code)
	if _, err := rw.Write(bs); err != nil {
		return fmt.Errorf("write to response failed: %w", err)
	}
	return nil
}
