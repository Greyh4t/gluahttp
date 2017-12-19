package gluahttp

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/yuin/gopher-lua"
)

func getResp(L *lua.LState, resp *http.Response) *lua.LTable {
	luaResp := makeResp(L, resp)
	luaResp.RawSetString("history", getHistory(L, resp))
	return luaResp
}

func makeResp(L *lua.LState, resp *http.Response) *lua.LTable {
	luaResp := L.NewTable()
	if resp != nil {
		luaResp.RawSetString("status_code", lua.LNumber(resp.StatusCode))
		body := getRespBody(resp)
		luaResp.RawSetString("body", lua.LString(body))
		luaResp.RawSetString("body_size", lua.LNumber(len(body)))
		luaResp.RawSetString("headers", getHeaders(L, resp.Header))
		luaResp.RawSetString("raw_headers", rawHeaders(resp.Header))
		luaResp.RawSetString("cookies", getCookies(L, resp.Cookies()))
		luaResp.RawSetString("raw_cookies", rawCookies(resp.Cookies()))
		luaResp.RawSetString("proto", lua.LString(resp.Proto))
		luaResp.RawSetString("url", lua.LString(resp.Request.URL.String()))
		luaResp.RawSetString("request", makeReq(L, resp.Request))
	}
	return luaResp
}

func makeReq(L *lua.LState, req *http.Request) *lua.LTable {
	luaReq := L.NewTable()
	if req != nil {
		luaReq.RawSetString("method", lua.LString(req.Method))
		luaReq.RawSetString("url", lua.LString(req.URL.String()))
		luaReq.RawSetString("scheme", lua.LString(req.URL.Scheme))
		luaReq.RawSetString("proto", lua.LString(req.Proto))
		luaReq.RawSetString("host", getHost(req))
		luaReq.RawSetString("body", lua.LString(getReqBody(req)))
		luaReq.RawSetString("headers", getHeaders(L, req.Header))
		luaReq.RawSetString("raw_headers", rawHeaders(req.Header))
		luaReq.RawSetString("cookies", getCookies(L, req.Cookies()))
		luaReq.RawSetString("raw_cookies", rawCookies(req.Cookies()))
		luaReq.RawSetString("raw", rawRequest(req))
	}
	return luaReq
}

func getHistory(L *lua.LState, resp *http.Response) *lua.LTable {
	history := L.NewTable()
	subResp := resp.Request.Response
	for {
		if subResp != nil {
			history.Insert(1, makeResp(L, subResp))
			subResp = subResp.Request.Response
		} else {
			break
		}
	}
	return history
}

func getHost(req *http.Request) lua.LString {
	if req.Host != "" {
		return lua.LString(req.Host)
	}
	return lua.LString(req.URL.Host)
}

func rawHeaders(headers http.Header) lua.LString {
	var rawHeader string
	for name, v := range headers {
		for _, vaule := range v {
			rawHeader += name + ": " + vaule + "\r\n"
		}
	}
	return lua.LString(strings.TrimSuffix(rawHeader, "\r\n"))
}

func getHeaders(L *lua.LState, headers http.Header) *lua.LTable {
	table := L.NewTable()
	for k, v := range headers {
		var each string
		for _, header := range v {
			each += header + ","
		}
		table.RawSetString(k, lua.LString(strings.TrimSuffix(each, ",")))
	}
	return table
}

func rawCookies(cookies []*http.Cookie) lua.LString {
	var rawCookie string
	for _, cookie := range cookies {
		rawCookie += cookie.Name + "=" + cookie.Value + ";"
	}
	return lua.LString(strings.TrimSuffix(rawCookie, ";"))
}

func getCookies(L *lua.LState, cookies []*http.Cookie) *lua.LTable {
	table := L.NewTable()
	for _, cookie := range cookies {
		table.RawSetString(cookie.Name, lua.LString(cookie.Value))
	}
	return table
}

func getRespBody(resp *http.Response) string {
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return string(body)
}

func getReqBody(req *http.Request) string {
	var body string
	if req.GetBody != nil {
		b, err := req.GetBody()
		if err == nil {
			buf, err := ioutil.ReadAll(b)
			b.Close()
			if err == nil {
				body = string(buf)
			}
		}
	}
	return body
}

func rawRequest(req *http.Request) lua.LString {
	rawRequest := req.Method + " " + req.URL.RequestURI() + " " + req.Proto + "\r\n"
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	rawRequest += "Host: " + host + "\r\n"
	for key, val := range req.Header {
		rawRequest += key + ": " + val[0] + "\r\n"
	}
	rawRequest += "\r\n" + getReqBody(req)
	return lua.LString(rawRequest)
}
