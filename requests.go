package gluahttp

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/yuin/gopher-lua"
)

var quoteEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

type fileUpload struct {
	// Filename is the name of the file that you wish to upload. We use this to guess the mimetype as well as pass it onto the server
	FileName string

	// FileContents is happy as long as you pass it a io.ReadCloser (which most file use anyways)
	FileContents io.ReadCloser

	// FieldName is form field name
	FieldName string

	// FileMime represents which mimetime should be sent along with the file.
	// When empty, defaults to application/octet-stream
	FileMime string
}

type requestOptions struct {

	// Data is a map of key values that will eventually convert into the
	// query string of a GET request or the body of a POST request.
	Data map[string]string

	// Params is a map of query strings that may be used within a GET request
	Params map[string]string

	// Files is where you can include files to upload. The use of this data
	// structure is limited to POST requests
	Files []fileUpload

	// JSON can be used when you wish to send JSON within the request body
	JSON string

	// XML can be used if you wish to send XML within the request body
	XML string

	// Headers if you want to add custom HTTP headers to the request,
	// this is your friend
	Headers map[string]string

	// InsecureSkipVerify is a flag that specifies if we should validate the
	// server's TLS certificate. It should be noted that Go's TLS verify mechanism
	// doesn't validate if a certificate has been revoked
	InsecureSkipVerify bool

	// DisableCompression will disable gzip compression on requests
	DisableCompression bool

	// Host allows you to set an arbitrary custom host
	Host string

	// Auth allows you to specify a user name and password that you wish to
	// use when requesting the URL. It will use basic HTTP authentication
	// formatting the username and password in base64 the format is:
	// []string{username, password}
	Auth []string

	// Cookies is an array of `http.Cookie` that allows you to attach
	// cookies to your request
	Cookies []*http.Cookie

	// Proxies is a map in the following format
	// *protocol* => proxy address e.g http => http://127.0.0.1:8080
	Proxies map[string]*url.URL

	// RequestTimeout is the maximum amount of time a whole request(include dial / request / redirect)
	// will wait.
	Timeout time.Duration

	DisableRedirect bool

	// RequestBody allows you to put anything matching an `io.Reader` into the request
	// this option will take precedence over any other request option specified
	//RequestBody io.Reader

	RawQuery string
	RawData  string
	IsAjax   bool
}

func (ro *requestOptions) CloseFiles() {
	for _, f := range ro.Files {
		f.FileContents.Close()
	}
}

func (ro requestOptions) proxySettings(req *http.Request) (*url.URL, error) {
	// No proxies – lets use the default
	if len(ro.Proxies) == 0 {
		return http.ProxyFromEnvironment(req)
	}

	// There was a proxy specified – do we support the protocol?
	if _, ok := ro.Proxies[req.URL.Scheme]; ok {
		return ro.Proxies[req.URL.Scheme], nil
	}

	// Proxies were specified but not for any protocol that we use
	return http.ProxyFromEnvironment(req)
}

func fileUploadFromDisk(fieldName, filePath string) (fileUpload, error) {
	var fu fileUpload

	fd, err := os.Open(filePath)

	if err != nil {
		return fu, err
	}
	_, fileName := path.Split(strings.Replace(filePath, `\`, `/`, -1))
	return fileUpload{
		FieldName:    fieldName,
		FileName:     fileName,
		FileContents: fd,
	}, nil
}

func parseOptions(options *lua.LTable) (*requestOptions, error) {
	var ro = new(requestOptions)
	if options == nil {
		ro.Timeout = 30 * time.Second
		return ro, nil
	}

	if reqFiles, ok := options.RawGetString("files").(*lua.LTable); ok {
		var err error
		reqFiles.ForEach(func(fieldName, filePath lua.LValue) {
			fu, ferr := fileUploadFromDisk(fieldName.String(), filePath.String())
			if ferr != nil {
				err = ferr
				return
			}
			ro.Files = append(ro.Files, fu)
		})
		if err != nil {
			return nil, err
		}
	}

	if reqProxies, ok := options.RawGetString("proxies").(*lua.LTable); ok {
		var err error
		ro.Proxies = map[string]*url.URL{}
		reqProxies.ForEach(func(scheme, proxy lua.LValue) {
			proxyUrl, perr := url.Parse(proxy.String())
			if perr != nil {
				err = perr
				return
			}
			ro.Proxies[scheme.String()] = proxyUrl
		})
		if err != nil {
			return nil, err
		}
	}

	if reqTimeout, ok := options.RawGetString("timeout").(lua.LNumber); ok {
		ro.Timeout = time.Duration(float64(lua.LVAsNumber(reqTimeout))) * time.Second
	}

	if reqVerify, ok := options.RawGetString("verify").(lua.LBool); ok {
		ro.InsecureSkipVerify = !bool(reqVerify)
	}

	if reqCompress, ok := options.RawGetString("compress").(lua.LBool); ok {
		ro.DisableCompression = !bool(reqCompress)
	}

	if reqAjax, ok := options.RawGetString("ajax").(lua.LBool); ok {
		ro.IsAjax = bool(reqAjax)
	}

	if reqRedirect, ok := options.RawGetString("redirect").(lua.LBool); ok {
		ro.DisableRedirect = !bool(reqRedirect)
	}

	if reqHost, ok := options.RawGetString("host").(lua.LString); ok {
		ro.Host = reqHost.String()
	}

	if reqAuth, ok := options.RawGetString("auth").(*lua.LTable); ok {
		ro.Auth = []string{
			reqAuth.RawGetInt(1).String(),
			reqAuth.RawGetInt(2).String(),
		}
	}

	if reqHeaders, ok := options.RawGetString("headers").(*lua.LTable); ok {
		ro.Headers = map[string]string{}
		reqHeaders.ForEach(func(key, value lua.LValue) {
			ro.Headers[key.String()] = value.String()
		})
	}

	if reqCookies, ok := options.RawGetString("cookies").(*lua.LTable); ok {
		reqCookies.ForEach(func(key, value lua.LValue) {
			ro.Cookies = append(ro.Cookies, &http.Cookie{Name: key.String(), Value: value.String()})
		})
	}

	if reqParams, ok := options.RawGetString("params").(*lua.LTable); ok {
		ro.Params = map[string]string{}
		reqParams.ForEach(func(key, value lua.LValue) {
			ro.Params[key.String()] = value.String()
		})
	}

	if reqData, ok := options.RawGetString("data").(*lua.LTable); ok {
		ro.Data = map[string]string{}
		reqData.ForEach(func(key, value lua.LValue) {
			ro.Data[key.String()] = value.String()
		})
	}

	if reqJson, ok := options.RawGetString("json").(lua.LString); ok {
		ro.JSON = reqJson.String()
	}

	if reqXml, ok := options.RawGetString("xml").(lua.LString); ok {
		ro.XML = reqXml.String()
	}

	if reqRawData, ok := options.RawGetString("raw_data").(lua.LString); ok {
		ro.RawData = reqRawData.String()
	}

	if reqQuery, ok := options.RawGetString("raw_query").(lua.LString); ok {
		ro.RawQuery = reqQuery.String()
	}

	return ro, nil
}

func (self *httpModule) createTransport(ro requestOptions) *http.Transport {
	transport := &http.Transport{
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		Proxy:              ro.proxySettings,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: ro.InsecureSkipVerify},
		DisableCompression: ro.DisableCompression,
	}

	if self.resolver != nil {
		transport.Dial = func(network, address string) (net.Conn, error) {
			host, port, _ := net.SplitHostPort(address)
			ip, err := self.resolver.FetchOneString(host)
			if err != nil {
				return nil, err
			}
			conn, err := net.DialTimeout(network, net.JoinHostPort(ip, port), ro.Timeout)
			if err != nil {
				return nil, err
			}
			return newTimeoutConn(conn, ro.Timeout), nil
		}
	} else {
		transport.Dial = func(network, address string) (net.Conn, error) {
			conn, err := net.DialTimeout(network, address, ro.Timeout)
			if err != nil {
				return nil, err
			}
			return newTimeoutConn(conn, ro.Timeout), nil
		}
	}

	ensureTransporterFinalized(transport)
	return transport
}

func (self *httpModule) buildClient(ro requestOptions) *http.Client {
	// The function does not return an error ever... so we are just ignoring it
	cookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	client := &http.Client{
		Jar:       cookieJar,
		Transport: self.createTransport(ro),
		Timeout:   ro.Timeout,
	}

	if ro.DisableRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}

func buildRequest(method, urlStr string, ro *requestOptions) (*http.Request, error) {
	if ro.RawData != "" {
		return http.NewRequest(method, urlStr, strings.NewReader(ro.RawData))
	}

	if ro.JSON != "" {
		return createBasicJSONRequest(method, urlStr, ro)
	}

	if ro.XML != "" {
		return createBasicXMLRequest(method, urlStr, ro)
	}

	if ro.Files != nil {
		return createFileUploadRequest(method, urlStr, ro)
	}

	if ro.Data != nil {
		return createBasicRequest(method, urlStr, ro)
	}

	return http.NewRequest(method, urlStr, nil)
}

func createBasicJSONRequest(method, urlStr string, ro *requestOptions) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, strings.NewReader(ro.JSON))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func createBasicXMLRequest(method, urlStr string, ro *requestOptions) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, strings.NewReader(ro.XML))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/xml")

	return req, nil
}

func createFileUploadRequest(method, urlStr string, ro *requestOptions) (*http.Request, error) {
	if method == "POST" {
		return createMultiPartPostRequest(method, urlStr, ro)
	}

	// This may be a PUT or PATCH request so we will just put the raw
	// io.ReadCloser in the request body
	// and guess the MIME type from the file name

	// At the moment, we will only support 1 file upload as a time
	// when uploading using PUT or PATCH

	req, err := http.NewRequest(method, urlStr, ro.Files[0].FileContents)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", mime.TypeByExtension(path.Ext(ro.Files[0].FileName)))

	return req, nil
}

func createMultiPartPostRequest(method, urlStr string, ro *requestOptions) (*http.Request, error) {
	body := &bytes.Buffer{}

	multipartWriter := multipart.NewWriter(body)

	for i, f := range ro.Files {
		if f.FileContents == nil {
			return nil, errors.New("Pointer FileContents cannot be nil")
		}

		fieldName := f.FieldName

		if fieldName == "" {
			if len(ro.Files) > 1 {
				fieldName = "file" + strconv.Itoa(i+1)
			} else {
				fieldName = "file"
			}
		}

		var writer io.Writer
		var err error

		if f.FileMime != "" {
			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes(fieldName), escapeQuotes(f.FileName)))
			h.Set("Content-Type", f.FileMime)
			writer, err = multipartWriter.CreatePart(h)
		} else {
			writer, err = multipartWriter.CreateFormFile(fieldName, f.FileName)
		}

		if err != nil {
			return nil, err
		}

		if _, err = io.Copy(writer, f.FileContents); err != nil && err != io.EOF {
			return nil, err
		}
	}

	// Populate the other parts of the form (if there are any)
	for key, value := range ro.Data {
		multipartWriter.WriteField(key, value)
	}

	if err := multipartWriter.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, urlStr, body)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", multipartWriter.FormDataContentType())

	return req, err
}

func createBasicRequest(method, urlStr string, ro *requestOptions) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, strings.NewReader(encodePostValues(ro.Data)))

	if err != nil {
		return nil, err
	}

	// The content type must be set to a regular form
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

func (self *httpModule) doRequest(L *lua.LState, method, urlStr string, options *lua.LTable) (lua.LValue, error) {
	ro, err := parseOptions(options)
	if err != nil {
		return lua.LNil, err
	}
	defer ro.CloseFiles()

	urlStr, err = buildURL(urlStr, ro)
	if err != nil {
		return lua.LNil, err
	}

	req, err := buildRequest(method, urlStr, ro)
	if err != nil {
		return lua.LNil, err
	}

	addHeaders(req, ro)
	addCookies(req, ro)

	client := self.buildClient(*ro)
	resp, err := client.Do(req)
	if err != nil {
		return lua.LNil, err
	}
	return getResp(L, resp), nil
}

// buildURLParams returns a URL with all of the params
// Note: This function will override current URL params if they contradict what is provided in the map
// That is what the "magic" is on the last line
func buildURL(urlStr string, ro *requestOptions) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	if ro.RawQuery != "" {
		parsedURL.RawQuery = ro.RawQuery
	} else {
		query := parsedURL.Query()
		if len(ro.Params) > 0 {
			for key, value := range ro.Params {
				query.Set(key, value)
			}
		}
		parsedURL.RawQuery = query.Encode()
	}

	return parsedURL.String(), nil
}

// addHTTPHeaders adds any additional HTTP headers that need to be added are added here including:
// 1. Authorization Headers
// 2. Any other header requested
func addHeaders(req *http.Request, ro *requestOptions) {
	req.Header.Set("X-SCANNER", "ZERO")

	for key, value := range ro.Headers {
		req.Header.Set(key, value)
	}

	if ro.Host != "" {
		req.Host = ro.Host
	}

	if ro.Auth != nil {
		req.SetBasicAuth(ro.Auth[0], ro.Auth[1])
	}

	if ro.IsAjax {
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}
}

func addCookies(req *http.Request, ro *requestOptions) {
	for _, c := range ro.Cookies {
		req.AddCookie(c)
	}
}

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func encodePostValues(postValues map[string]string) string {
	urlValues := &url.Values{}

	for key, value := range postValues {
		urlValues.Set(key, value)
	}

	return urlValues.Encode() // This will sort all of the string values
}

// ensureTransporterFinalized will ensure that when the HTTP client is GCed
// the runtime will close the idle connections (so that they won't leak)
// this function was adopted from Hashicorp's go-cleanhttp package
func ensureTransporterFinalized(httpTransport *http.Transport) {
	runtime.SetFinalizer(&httpTransport, func(transportInt **http.Transport) {
		(*transportInt).CloseIdleConnections()
	})
}
