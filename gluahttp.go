package gluahttp

import (
	"github.com/Greyh4t/dnscache"
	"github.com/yuin/gopher-lua"
)

type httpModule struct {
	resolver *dnscache.Resolver
}

func NewHttpModule(resolver *dnscache.Resolver) *httpModule {
	return &httpModule{
		resolver: resolver,
	}
}

func (self *httpModule) Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"get":     self.get,
		"delete":  self.delete,
		"head":    self.head,
		"patch":   self.patch,
		"post":    self.post,
		"put":     self.put,
		"options": self.options,
	})
	L.Push(mod)
	return 1
}

func (self *httpModule) get(L *lua.LState) int {
	return self.doRequestAndPush(L, "GET", L.ToString(1), L.ToTable(2))
}

func (self *httpModule) delete(L *lua.LState) int {
	return self.doRequestAndPush(L, "DELETE", L.ToString(1), L.ToTable(2))
}

func (self *httpModule) head(L *lua.LState) int {
	return self.doRequestAndPush(L, "HEAD", L.ToString(1), L.ToTable(2))
}

func (self *httpModule) patch(L *lua.LState) int {
	return self.doRequestAndPush(L, "PATCH", L.ToString(1), L.ToTable(2))
}

func (self *httpModule) post(L *lua.LState) int {
	return self.doRequestAndPush(L, "POST", L.ToString(1), L.ToTable(2))
}

func (self *httpModule) put(L *lua.LState) int {
	return self.doRequestAndPush(L, "PUT", L.ToString(1), L.ToTable(2))
}

func (self *httpModule) options(L *lua.LState) int {
	return self.doRequestAndPush(L, "OPTIONS", L.ToString(1), L.ToTable(2))
}

func (self *httpModule) doRequestAndPush(L *lua.LState, method string, uri string, options *lua.LTable) int {
	response, err := self.doRequest(L, method, uri, options)

	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(response)
	return 1
}
