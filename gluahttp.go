package gluahttp

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/gopher-lua"
)

type httpModule struct {
	client *http.Client
}

func NewHttpModule() *httpModule {
	return &httpModule{
		client: &http.Client{},
	}
}

type empty struct{}

func (h *httpModule) Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"get":     h.get,
		"delete":  h.delete,
		"head":    h.head,
		"patch":   h.patch,
		"post":    h.post,
		"put":     h.put,
		"options": h.options,
	})
	registerHttpResponseType(mod, L)
	L.Push(mod)
	return 1
}

func (h *httpModule) get(L *lua.LState) int {
	return h.doRequestAndPush(L, "GET", L.ToString(1), L.ToTable(2))
}

func (h *httpModule) delete(L *lua.LState) int {
	return h.doRequestAndPush(L, "DELETE", L.ToString(1), L.ToTable(2))
}

func (h *httpModule) head(L *lua.LState) int {
	return h.doRequestAndPush(L, "HEAD", L.ToString(1), L.ToTable(2))
}

func (h *httpModule) patch(L *lua.LState) int {
	return h.doRequestAndPush(L, "PATCH", L.ToString(1), L.ToTable(2))
}

func (h *httpModule) post(L *lua.LState) int {
	return h.doRequestAndPush(L, "POST", L.ToString(1), L.ToTable(2))
}

func (h *httpModule) put(L *lua.LState) int {
	return h.doRequestAndPush(L, "PUT", L.ToString(1), L.ToTable(2))
}

func (h *httpModule) options(L *lua.LState) int {
	return h.doRequestAndPush(L, "OPTIONS", L.ToString(1), L.ToTable(2))
}

func newfileUploadRequest(method, uri string, data *lua.LTable, files *lua.LTable) (*http.Request, error) {
	var (
		f   *os.File
		err error
	)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if files != nil {
		files.ForEach(func(name, filePath lua.LValue) {
			f, err = os.Open(filePath.String())
			if err == nil {
				part, err := writer.CreateFormFile(name.String(), filepath.Base(filePath.String()))
				if err == nil {
					_, err = io.Copy(part, f)
				}
				f.Close()
			}
		})
	}
	if err != nil {
		return nil, err
	}

	if data != nil {
		data.ForEach(func(key, value lua.LValue) {
			writer.WriteField(key.String(), value.String())
		})
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func (h *httpModule) doRequest(L *lua.LState, method string, uri string, options *lua.LTable) (*lua.LUserData, error) {
	var (
		req *http.Request
		err error
	)
	if options != nil {
		transport := &http.Transport{}
		if reqVerify, ok := options.RawGet(lua.LString("verifycert")).(lua.LBool); ok {
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: !bool(reqVerify)}
		}

		if reqTimeout, ok := options.RawGet(lua.LString("timeout")).(lua.LNumber); ok {
			timeout := time.Second * time.Duration(float64(lua.LVAsNumber(reqTimeout)))
			transport.Dial = func(network, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(network, addr, timeout)
				if err != nil {
					return nil, err
				}
				conn.SetDeadline(time.Now().Add(timeout))
				return conn, nil
			}
		}

		if reqProxy, ok := options.RawGet(lua.LString("proxy")).(lua.LString); ok {
			if reqProxy.String() == "" {
				transport.Proxy = nil
			} else {
				parsedProxyUrl, err := url.Parse(reqProxy.String())
				if err != nil {
					return nil, err
				}
				transport.Proxy = http.ProxyURL(parsedProxyUrl)
			}
		}

		if reqRedirect, ok := options.RawGet(lua.LString("redirect")).(lua.LBool); ok {
			if !bool(reqRedirect) {
				h.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}
			}
		}

		h.client.Transport = transport

		if reqFiles, ok := options.RawGet(lua.LString("files")).(*lua.LTable); ok {
			if reqData, ok := options.RawGet(lua.LString("data")).(*lua.LTable); ok {
				req, err = newfileUploadRequest(method, uri, reqData, reqFiles)
			} else {
				req, err = newfileUploadRequest(method, uri, nil, reqFiles)
			}
		} else if reqData, ok := options.RawGet(lua.LString("data")).(*lua.LTable); ok {
			urlValues := &url.Values{}
			reqData.ForEach(func(key, value lua.LValue) {
				urlValues.Set(key.String(), value.String())
			})
			req, err = http.NewRequest(method, uri, strings.NewReader(urlValues.Encode()))
			if err == nil {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
		} else if reqRawData, ok := options.RawGet(lua.LString("rawdata")).(lua.LString); ok {
			req, err = http.NewRequest(method, uri, strings.NewReader(reqRawData.String()))
		} else if reqJson, ok := options.RawGet(lua.LString("json")).(lua.LString); ok {
			req, err = http.NewRequest(method, uri, strings.NewReader(reqJson.String()))
			if err == nil {
				req.Header.Set("Content-Type", "application/json")
			}
		} else {
			req, err = http.NewRequest(method, uri, nil)
		}
		if err != nil {
			return nil, err
		}

		if reqHeaders, ok := options.RawGet(lua.LString("headers")).(*lua.LTable); ok {
			reqHeaders.ForEach(func(key, value lua.LValue) {
				req.Header.Set(key.String(), value.String())
			})
		}
		if reqCookies, ok := options.RawGet(lua.LString("cookies")).(*lua.LTable); ok {
			reqCookies.ForEach(func(key lua.LValue, value lua.LValue) {
				req.AddCookie(&http.Cookie{Name: key.String(), Value: value.String()})
			})
		}
		if reqParams, ok := options.RawGet(lua.LString("params")).(*lua.LTable); ok {
			parsedQuery := req.URL.Query()
			reqParams.ForEach(func(key, value lua.LValue) {
				if _, ok := parsedQuery[key.String()]; ok {
					parsedQuery.Add(key.String(), value.String())
					return
				}
				parsedQuery.Set(key.String(), value.String())
			})
			req.URL.RawQuery = parsedQuery.Encode()
		}
		if reqHost, ok := options.RawGet(lua.LString("host")).(lua.LString); ok {
			req.Host = reqHost.String()
		}

		if reqBasicAuth, ok := options.RawGet(lua.LString("basicauth")).(*lua.LTable); ok {
			req.SetBasicAuth(reqBasicAuth.RawGetInt(0).String(), reqBasicAuth.RawGetInt(1).String())
		}
	}
	req.Close = true
	res, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	return newHttpResponse(res, &body, len(body), L), nil
}

func (h *httpModule) doRequestAndPush(L *lua.LState, method string, uri string, options *lua.LTable) int {
	response, err := h.doRequest(L, method, uri, options)

	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("%s", err)))
		return 2
	}

	L.Push(response)
	return 1
}

func toTable(v lua.LValue) *lua.LTable {
	if lv, ok := v.(*lua.LTable); ok {
		return lv
	}
	return nil
}
