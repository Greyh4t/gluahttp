package gluahttp

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path/filepath"
	"strings"
	"time"

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

func (self *httpModule) newfileUploadRequest(method, u string, data *lua.LTable, files *lua.LTable) (*http.Request, error) {
	var body *bytes.Buffer
	writer := multipart.NewWriter(body)

	var fileList [][]string
	if files != nil {
		files.ForEach(func(name, path lua.LValue) {
			fileList = append(fileList, []string{name.String(), path.String()})
		})
	}

	for _, file := range fileList {
		buf, err := ioutil.ReadFile(file[1])
		if err != nil {
			return nil, err
		}
		part, err := writer.CreateFormFile(file[0], filepath.Base(file[1]))
		if err != nil {
			return nil, err
		}
		_, err = part.Write(buf)
		if err != nil {
			return nil, err
		}
	}

	if data != nil {
		data.ForEach(func(key, value lua.LValue) {
			writer.WriteField(key.String(), value.String())
		})
	}
	err := writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, u, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func (self *httpModule) doRequest(L *lua.LState, method, u string, options *lua.LTable) (lua.LValue, error) {
	var (
		req       *http.Request
		client    = new(http.Client)
		transport = new(http.Transport)
		err       error
	)
	client.Jar, _ = cookiejar.New(nil)
	transport.MaxIdleConns = 1000

	if options != nil {
		if rawUrl, _ := options.RawGetString("rawquery").(lua.LBool); !rawUrl {
			parsedUrl, err := url.Parse(u)
			if err != nil {
				return lua.LNil, err
			}
			parsedUrl.RawQuery = strings.Replace(parsedUrl.Query().Encode(), "+", "%20", -1)
			u = parsedUrl.String()
		}

		if reqProxy, ok := options.RawGetString("proxy").(lua.LString); ok {
			if reqProxy.String() == "" {
				transport.Proxy = nil
			} else {
				parsedProxyUrl, err := url.Parse(reqProxy.String())
				if err != nil {
					return lua.LNil, err
				}
				transport.Proxy = http.ProxyURL(parsedProxyUrl)
			}
		}

		if reqRedirect, ok := options.RawGetString("redirect").(lua.LBool); ok {
			if !bool(reqRedirect) {
				client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}
			}
		}

		if reqVerify, ok := options.RawGetString("verifycert").(lua.LBool); ok {
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: !bool(reqVerify)}
		}

		if reqTimeout, ok := options.RawGetString("timeout").(lua.LNumber); ok {
			timeout := time.Second * time.Duration(float64(lua.LVAsNumber(reqTimeout)))
			client.Timeout = timeout
			transport.IdleConnTimeout = timeout
			transport.TLSHandshakeTimeout = timeout
			if self.resolver != nil {
				transport.Dial = func(network string, address string) (net.Conn, error) {
					host, port, _ := net.SplitHostPort(address)
					ip, err := self.resolver.FetchOneString(host)
					if err != nil {
						return nil, err
					}
					return net.DialTimeout("tcp", net.JoinHostPort(ip, port), timeout)
				}
			} else {
				transport.Dial = (&net.Dialer{
					Timeout: timeout,
				}).Dial
			}
		}

		//make request
		if reqFiles, ok := options.RawGetString("files").(*lua.LTable); ok {
			if reqData, ok := options.RawGetString("data").(*lua.LTable); ok {
				req, err = self.newfileUploadRequest(method, u, reqData, reqFiles)
			} else {
				req, err = self.newfileUploadRequest(method, u, nil, reqFiles)
			}
		} else if reqData, ok := options.RawGetString("data").(*lua.LTable); ok {
			urlValues := &url.Values{}
			reqData.ForEach(func(key, value lua.LValue) {
				urlValues.Set(key.String(), value.String())
			})
			req, err = http.NewRequest(method, u, strings.NewReader(urlValues.Encode()))
			if err == nil {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
		} else if reqRawData, ok := options.RawGetString("rawdata").(lua.LString); ok {
			req, err = http.NewRequest(method, u, strings.NewReader(reqRawData.String()))
		} else if reqJson, ok := options.RawGetString("json").(lua.LString); ok {
			req, err = http.NewRequest(method, u, strings.NewReader(reqJson.String()))
			if err == nil {
				req.Header.Set("Content-Type", "application/json")
			}
		} else {
			req, err = http.NewRequest(method, u, nil)
		}
		if err != nil {
			return lua.LNil, err
		}

		if reqHeaders, ok := options.RawGetString("headers").(*lua.LTable); ok {
			reqHeaders.ForEach(func(key, value lua.LValue) {
				req.Header.Set(key.String(), value.String())
			})
		}

		if reqCookies, ok := options.RawGetString("cookies").(*lua.LTable); ok {
			reqCookies.ForEach(func(key lua.LValue, value lua.LValue) {
				req.AddCookie(&http.Cookie{Name: key.String(), Value: value.String()})
			})
		}

		if reqParams, ok := options.RawGetString("params").(*lua.LTable); ok {
			parsedQuery := req.URL.Query()
			reqParams.ForEach(func(key, value lua.LValue) {
				if _, ok := parsedQuery[key.String()]; ok {
					parsedQuery.Add(key.String(), value.String())
					return
				}
				parsedQuery.Set(key.String(), value.String())
			})
			if rawUrl, _ := options.RawGetString("rawquery").(lua.LBool); !rawUrl {
				req.URL.RawQuery = strings.Replace(parsedQuery.Encode(), "+", "%20", -1)
			} else {
				rawQuery, _ := url.QueryUnescape(parsedQuery.Encode())
				req.URL.RawQuery = rawQuery
			}
		}

		if reqHost, ok := options.RawGetString("host").(lua.LString); ok {
			req.Host = reqHost.String()
		}

		if reqBasicAuth, ok := options.RawGetString("basicauth").(*lua.LTable); ok {
			req.SetBasicAuth(reqBasicAuth.RawGetInt(1).String(), reqBasicAuth.RawGetInt(2).String())
		}
	} else {
		req, err = http.NewRequest(method, u, nil)
		if err != nil {
			return lua.LNil, err
		}
		req.Close = true

		client.Timeout = time.Second * 30
		transport.IdleConnTimeout = time.Second * 10
		transport.TLSHandshakeTimeout = time.Second * 10

		if self.resolver != nil {
			transport.Dial = func(network string, address string) (net.Conn, error) {
				host, port, _ := net.SplitHostPort(address)
				ip, err := self.resolver.FetchOneString(host)
				if err != nil {
					return nil, err
				}
				return net.DialTimeout("tcp", net.JoinHostPort(ip, port), time.Second*10)
			}
		} else {
			transport.Dial = (&net.Dialer{
				Timeout: time.Second * 10,
			}).Dial
		}
	}

	client.Transport = transport

	resp, err := client.Do(req)
	if err != nil {
		return lua.LNil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		return lua.LNil, err
	}

	return makeResp(L, resp, string(body)), nil
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
