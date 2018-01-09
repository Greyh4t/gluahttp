package gluahttp

import (
	"github.com/Greyh4t/dnscache"
	"github.com/yuin/gopher-lua"
)

type httpModule struct {
	resolver *dnscache.Resolver
}

func New(resolver *dnscache.Resolver) *httpModule {
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

func (self *httpModule) AsyncLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"get":     self.asyncGet,
		"delete":  self.asyncPost,
		"head":    self.asyncHead,
		"patch":   self.asyncPatch,
		"post":    self.asyncPost,
		"put":     self.asyncPut,
		"options": self.asyncOptions,
	})
	L.Push(mod)
	return 1
}

//sync
func (self *httpModule) get(L *lua.LState) int {
	return self.doRequestAndPush(L, "GET", L.CheckString(1), L.ToTable(2))
}

func (self *httpModule) delete(L *lua.LState) int {
	return self.doRequestAndPush(L, "DELETE", L.CheckString(1), L.ToTable(2))
}

func (self *httpModule) head(L *lua.LState) int {
	return self.doRequestAndPush(L, "HEAD", L.CheckString(1), L.ToTable(2))
}

func (self *httpModule) patch(L *lua.LState) int {
	return self.doRequestAndPush(L, "PATCH", L.CheckString(1), L.ToTable(2))
}

func (self *httpModule) post(L *lua.LState) int {
	return self.doRequestAndPush(L, "POST", L.CheckString(1), L.ToTable(2))
}

func (self *httpModule) put(L *lua.LState) int {
	return self.doRequestAndPush(L, "PUT", L.CheckString(1), L.ToTable(2))
}

func (self *httpModule) options(L *lua.LState) int {
	return self.doRequestAndPush(L, "OPTIONS", L.CheckString(1), L.ToTable(2))
}

func (self *httpModule) doRequestAndPush(L *lua.LState, method string, url string, options *lua.LTable) int {
	response, err := self.doRequest(L, method, url, options)

	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(response)
	return 1
}

//async
func (self *httpModule) asyncGet(L *lua.LState) int {
	return self.asyncDoRequestAndPush(L, "GET", L.CheckString(1), L.ToTable(2))
}

func (self *httpModule) asyncDelete(L *lua.LState) int {
	return self.asyncDoRequestAndPush(L, "DELETE", L.CheckString(1), L.ToTable(2))
}

func (self *httpModule) asyncHead(L *lua.LState) int {
	return self.asyncDoRequestAndPush(L, "HEAD", L.CheckString(1), L.ToTable(2))
}

func (self *httpModule) asyncPatch(L *lua.LState) int {
	return self.asyncDoRequestAndPush(L, "PATCH", L.CheckString(1), L.ToTable(2))
}

func (self *httpModule) asyncPost(L *lua.LState) int {
	return self.asyncDoRequestAndPush(L, "POST", L.CheckString(1), L.ToTable(2))
}

func (self *httpModule) asyncPut(L *lua.LState) int {
	return self.asyncDoRequestAndPush(L, "PUT", L.CheckString(1), L.ToTable(2))
}

func (self *httpModule) asyncOptions(L *lua.LState) int {
	return self.asyncDoRequestAndPush(L, "OPTIONS", L.CheckString(1), L.ToTable(2))
}

func (self *httpModule) asyncDoRequestAndPush(L *lua.LState, method string, url string, options *lua.LTable) int {
	resultChan := make(chan lua.LValue, 2)

	go func(L *lua.LState, method string, url string, options *lua.LTable, resultChan chan lua.LValue) {
		response, err := self.doRequest(L, method, url, options)
		if err != nil {
			resultChan <- lua.LNil
			resultChan <- lua.LString(err.Error())
		} else {
			resultChan <- response
		}
		close(resultChan)
	}(L, method, url, options, resultChan)

	return L.Yield(lua.LChannel(resultChan))
}
