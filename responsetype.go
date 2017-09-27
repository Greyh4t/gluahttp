package gluahttp

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/yuin/gopher-lua"
)

func makeResp(L *lua.LState, resp *http.Response, body string) lua.LValue {
	luaResp := L.NewTable()
	luaResp.RawSetString("status_code", lua.LNumber(resp.StatusCode))
	luaResp.RawSetString("body", lua.LString(body))
	luaResp.RawSetString("body_size", lua.LNumber(len(body)))
	luaResp.RawSetString("header", header(L, resp))
	luaResp.RawSetString("raw_header", rawHeader(resp))
	luaResp.RawSetString("cookie", cookie(L, resp))
	luaResp.RawSetString("raw_cookie", rawCookie(resp))
	luaResp.RawSetString("url", lua.LString(resp.Request.URL.String()))
	luaResp.RawSetString("req_scheme", lua.LString(resp.Request.URL.Scheme))
	luaResp.RawSetString("raw_req", rawRequest(resp))
	luaResp.RawSetString("proto", lua.LString(resp.Proto))
	return luaResp
}

func rawHeader(resp *http.Response) lua.LValue {
	var rawHeader string
	for name, v := range resp.Header {
		for _, vaule := range v {
			rawHeader += name + ": " + vaule + "\r\n"
		}
	}
	return lua.LString(strings.TrimSuffix(rawHeader, "\r\n"))
}

func header(L *lua.LState, resp *http.Response) lua.LValue {
	table := L.NewTable()
	for k, v := range resp.Header {
		var each string
		for _, header := range v {
			each += header + ", "
		}
		table.RawSetString(k, lua.LString(strings.TrimSuffix(each, ", ")))
	}
	return table
}

func rawCookie(resp *http.Response) lua.LValue {
	var rawCookie string
	for _, cookie := range resp.Cookies() {
		rawCookie += cookie.Name + "=" + cookie.Value + ";"
	}
	return lua.LString(strings.TrimSuffix(rawCookie, ";"))
}

func cookie(L *lua.LState, resp *http.Response) lua.LValue {
	table := L.NewTable()
	for _, cookie := range resp.Cookies() {
		table.RawSetString(cookie.Name, lua.LString(cookie.Value))
	}
	return table
}

func rawRequest(resp *http.Response) lua.LValue {
	r := resp.Request
	rawRequest := r.Method + " " + r.URL.RequestURI() + " " + r.Proto + "\r\n"
	host := r.Host
	if host == "" {
		host = r.URL.Host
	}
	rawRequest += "Host: " + host + "\r\n"
	for key, val := range r.Header {
		rawRequest += key + ": " + val[0] + "\r\n"
	}
	rawRequest += "\r\n"
	if r.GetBody != nil {
		body, err := r.GetBody()
		if err == nil {
			buf, err := ioutil.ReadAll(body)
			body.Close()
			if err == nil {
				rawRequest += string(buf)
			}
		}
	}
	return lua.LString(rawRequest)
}
