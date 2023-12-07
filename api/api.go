package api

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hudangwei/common/macaron"
	"github.com/hudangwei/common/macaron/middleware"
)

const (
	APITagGuest     = "guest"
	APITagTeam      = "team"
	APITagDeveloper = "developer"
)

type APIManager interface {
	AddAPI(tag, method, path string)
	GetAPIs(tag string) map[string]string
}

type APIs struct {
	apis       map[string]string
	prefixApis map[string]struct{}
	taggedApis map[string]map[string]string
}

func NewAPIs() *APIs {
	return &APIs{
		apis:       make(map[string]string),
		prefixApis: make(map[string]struct{}),
		taggedApis: make(map[string]map[string]string),
	}
}

func ApiHash(api string) string {
	m := md5.Sum([]byte(api))
	return hex.EncodeToString(m[:])[8:24]
}

func (s *APIs) AddAPI(tag, method, path string) {
	key := fmt.Sprintf("%s %s", strings.ToLower(method), strings.ToLower(path))
	hash := ApiHash(key)
	s.apis[hash] = key
	if strings.HasSuffix(key, "*") {
		s.prefixApis[strings.ReplaceAll(key, "/*", "")] = struct{}{}
	}
	if len(tag) > 0 {
		apiSet, ok := s.taggedApis[tag]
		if !ok {
			apiSet = make(map[string]string)
			s.taggedApis[tag] = apiSet
		}
		apiSet[hash] = key
	}
}

func (s *APIs) GetAPIs(tag string) map[string]string {
	return s.taggedApis[tag]
}

var DefaultAPIManager APIManager = NewAPIs()
var DefaultMacaron *macaron.Macaron = NewMacaron()

func NewMacaron() *macaron.Macaron {
	m := macaron.New()
	m.Map(middleware.HTTPResp())
	m.Use(middleware.Parse())
	return m
}

func Handle(tag string, group *gin.RouterGroup, method, path string, handler macaron.Handler) {
	DefaultAPIManager.AddAPI(tag, method, path)
	group.Handle(method, path, DefaultMacaron.Wraps(handler))
}
