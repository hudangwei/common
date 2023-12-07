package middleware

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/astaxie/beego/validation"
	"github.com/hudangwei/common/logger"
	"github.com/hudangwei/common/macaron"
	"github.com/hudangwei/common/macaron/codec"
	"github.com/hudangwei/common/macaron/util"
	"go.uber.org/zap"
)

var Codec = []codec.Interface{
	&codec.Json{},
	&codec.MultipartForm{},
	&codec.Empty{},
}

func Parse() macaron.Handler {
	return func(ctx *macaron.Context) int {
		if ctx.InputType == nil {
			return 0
		}
		isReadFromBody := false
		switch ctx.Req.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			isReadFromBody = true
		}
		pInput := reflect.New(ctx.InputType).Interface()
		searchMap, err := injectFieldFromBody(pInput, isReadFromBody, ctx.Req)
		if err != nil {
			logger.Error("injectFieldFromBody with error", zap.Error(err))
			return macaron.Abort //解析数据失败，直接中止后续handler链
		}
		if err := injectFieldFromUrlAndMap(pInput, isReadFromBody, ctx, searchMap); err != nil {
			logger.Error("injectFieldFromUrlAndMap with error", zap.Error(err))
			return macaron.Abort
		}
		valid := validation.Validation{}
		ok, err := valid.Valid(pInput)
		if err != nil {
			logger.Error("valid with error", zap.Error(err))
			return macaron.Abort
		}
		if !ok {
			for _, err := range valid.Errors {
				logger.Warn("valid with error", zap.String("key", err.Key), zap.String("errmsg", err.Message))
			}
			return macaron.Abort
		}
		ctx.Map(pInput)
		return 0
	}
}

func injectFieldFromBody(ptr interface{}, isReadFromBody bool, req *http.Request) (map[string][]byte, error) {
	if !isReadFromBody || req.ContentLength == 0 {
		return nil, nil
	}

	contentType := req.Header.Get("Content-Type")
	var coc codec.Interface = &codec.Json{}
	for _, c := range Codec {
		for _, ctt := range c.ContentType() {
			if strings.HasPrefix(contentType, ctt) {
				coc = c
			}
		}
	}

	if dir, ok := coc.(codec.Direct); ok {
		if err := dir.Unmarshal(req, ptr); err != nil {
			return nil, err
		}
	}

	if s, ok := coc.(codec.Search); ok {
		m, err := s.UnmarshalSearchMap(req)
		if err != nil {
			return nil, err
		}
		return m, nil
	}

	return nil, nil
}

func injectFieldFromUrlAndMap(ptr interface{}, isReadFromBody bool, ctx *macaron.Context, searchMap map[string][]byte) error {
	elType := reflect.TypeOf(ptr).Elem()
	input := reflect.ValueOf(ptr).Elem()

	for i := 0; i < elType.NumField(); i++ {
		if input.Field(i).Kind() == reflect.Struct {
			if err := injectFieldFromUrlAndMap(input.Field(i).Addr().Interface(), isReadFromBody, ctx, searchMap); err != nil {
				return err
			}
			continue
		}

		src, name := getSourceWayAndName(elType.Field(i))
		if src == "" && isReadFromBody {
			if searchMap != nil {
				if v, ok := searchMap[name]; ok {
					if input.Field(i).Kind() == reflect.String {
						input.Field(i).Set(reflect.ValueOf(string(v)))
					} else if input.Field(i).Kind() == reflect.Slice {
						input.Field(i).Set(reflect.ValueOf(v))
					}
				}
			}
			if name == "@body" {
				bs, err := util.CopyBody(ctx.Req)
				if err != nil {
					return err
				}
				input.Field(i).Set(reflect.ValueOf(bs))
			}
			continue
		}

		val := ""
		switch src {
		case "path":
			val = ctx.PathParamsFunc(name)
		case "header":
			val = ctx.Req.Header.Get(name)
		default:
			val = ctx.Req.FormValue(name)
		}

		tarVal, err := changeToFieldKind(val, input.Field(i).Type())
		if err != nil {
			return err
		}
		if tarVal != nil {
			input.Field(i).Set(reflect.ValueOf(tarVal))
		}
	}

	return nil
}

func getSourceWayAndName(field reflect.StructField) (src, name string) {
	src, name = "", lowerFirst(field.Name)
	tag := field.Tag.Get("auto_read")
	if tag == "" {
		return
	}

	tagArr := strings.Split(tag, ",")
	name = strings.TrimSpace(tagArr[0])
	if len(tagArr) > 1 {
		src = strings.TrimSpace(tagArr[1])
	}

	return
}

func lowerFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

func changeToFieldKind(str string, t reflect.Type) (interface{}, error) {
	kind := t.Kind()
	isPtr := false
	if kind == reflect.Ptr {
		if str == "" {
			return nil, nil
		}
		isPtr = true
		kind = t.Elem().Kind()
	}

	if kind == reflect.String {
		if isPtr {
			return &str, nil
		}
		return str, nil
	}

	if kind == reflect.Bool {
		if str == "" {
			return false, nil
		}
		b, err := strconv.ParseBool(str)
		if err != nil {
			return nil, fmt.Errorf("changeToFieldKind covert to bool failed: %s", err)
		}
		if isPtr {
			return &b, nil
		}
		return b, nil
	}

	if kind == reflect.Int8 {
		if str == "" {
			return int8(0), nil
		}
		i, err := strconv.ParseInt(str, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("changeToFieldKind covert to int failed: %s", err)
		}

		i8 := int8(i)
		if isPtr {
			return &i8, nil
		}
		return i8, nil
	}

	if kind == reflect.Uint8 {
		if str == "" {
			return uint8(0), nil
		}
		i, err := strconv.ParseInt(str, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("changeToFieldKind covert to int failed: %s", err)
		}

		u8 := uint8(i)
		if isPtr {
			return &u8, nil
		}
		return u8, nil
	}

	if kind == reflect.Int {
		if str == "" {
			return int(0), nil
		}
		i, err := strconv.ParseInt(str, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("changeToFieldKind covert to int failed: %s", err)
		}

		i32 := int(i)
		if isPtr {
			return &i32, nil
		}
		return i32, nil
	}

	if kind == reflect.Uint {
		if str == "" {
			return uint(0), nil
		}
		i, err := strconv.ParseUint(str, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("changeToFieldKind covert to uint failed: %s", err)
		}

		ui := uint(i)
		if isPtr {
			return &ui, nil
		}
		return ui, nil
	}

	if kind == reflect.Int64 {
		if str == "" {
			return int64(0), nil
		}
		i, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("changeToFieldKind covert to int64 failed: %s", err)
		}

		if isPtr {
			return &i, nil
		}
		return i, nil
	}

	if kind == reflect.Uint64 {
		if str == "" {
			return uint64(0), nil
		}
		i, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("changeToFieldKind covert to uint64 failed: %s", err)
		}

		if isPtr {
			return &i, nil
		}
		return i, nil
	}

	return nil, fmt.Errorf("unsupport type: %s", kind.String())
}
