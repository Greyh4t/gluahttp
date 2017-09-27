# gluahttp

[![](https://travis-ci.org/cjoudrey/gluahttp.svg)](https://travis-ci.org/cjoudrey/gluahttp)

gluahttp provides an easy way to make HTTP requests from within [GopherLua](https://github.com/yuin/gopher-lua).

## Installation

```
go get github.com/cjoudrey/gluahttp
```

## Usage

```go
package main

import "github.com/yuin/gopher-lua"
import "github.com/Greyh4t/gluahttp"

func main() {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", gluahttp.Loader)

	if err := L.DoString(`
local http = require("http")
resp, err = http.post("http://www.example.com?x=1", {
	params={
		page="测试",
		x="2"
	},
	data={
		a=1,
		b="2"
	},
	--basicauth={"user", "passwd"},
	--files={
	--	file="/x.txt"
	--},
	--json='{"a":"测试"}',
	--rawdata="aaa'gdgd",
	headers={
		Accept="*/*",
		Test="xxx",
		["User-Agent"]="testUserAgent",
	},
	cookies={
		session="xxx",
		user="test"
	},
	--host="www.test.com",
	--proxy="http://127.0.0.1:8080",
	timeout=3,
	redirect=false,
	verifycert=false
})
if err
then
	print(err)
else
	print(resp.status_code)
	print("--------------")
	print(resp.body_size)
	print("--------------")
	print(resp.body)
	print("--------------")
	print(resp.header)
	print("--------------")
	print(resp.header["Content-Type"])
	print("--------------")
	print(resp.raw_header)
	print("--------------")
	print(resp.cookie)
	print("--------------")
	print(resp.raw_cookie)
	print("--------------")
	print(resp.url)
	print("--------------")
	print(resp.req_scheme)
	print("--------------")
	print(resp.proto)
	print("--------------")
	print(resp.raw_req)
end
	`); err != nil {
		panic(err)
	}
}
```

## API

- [`http.delete(url [, options])`](#httpdeleteurl--options)
- [`http.get(url [, options])`](#httpgeturl--options)
- [`http.head(url [, options])`](#httpheadurl--options)
- [`http.patch(url [, options])`](#httppatchurl--options)
- [`http.post(url [, options])`](#httpposturl--options)
- [`http.put(url [, options])`](#httpputurl--options)
- [`http.options(url [, options])`](#httpputurl--options)
- [`http.response`](#httpresponse)

### http.delete(url [, options])

**Attributes**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| url     | String | URL of the resource to load |
| options | Table  | Additional options |

**Options**

| Name       | Type    | Description |
| ---------- | ------- | ----------- |
| params     | Table   | URL encoded query params |
| cookies    | Table   | Additional cookies to send with the request |
| headers    | Table   | Additional headers to send with the request |
| proxy      | String  | Proxy |
| timeout    | Float64 | Dial timeout |
| redirect   | Bool    | Whether follow redirect |
| verifycert | Bool    | Whether verify server cert |
| host       | String  | Set a host different with url |

**Returns**

[http.response](#httpresponse) or (nil, error message)

### http.get(url [, options])

**Attributes**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| url     | String | URL of the resource to load |
| options | Table  | Additional options |

**Options**

| Name       | Type    | Description |
| ---------- | ------- | ----------- |
| params     | Table   | URL encoded query params |
| cookies    | Table   | Additional cookies to send with the request |
| headers    | Table   | Additional headers to send with the request |
| proxy      | String  | Proxy |
| timeout    | Float64 | Dial timeout |
| redirect   | Bool    | Whether follow redirect |
| verifycert | Bool    | Whether verify server cert |
| host       | String  | Set a host different with url |

**Returns**

[http.response](#httpresponse) or (nil, error message)

### http.head(url [, options])

**Attributes**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| url     | String | URL of the resource to load |
| options | Table  | Additional options |

**Options**

| Name       | Type    | Description |
| ---------- | ------- | ----------- |
| params     | Table   | URL encoded query params |
| cookies    | Table   | Additional cookies to send with the request |
| headers    | Table   | Additional headers to send with the request |
| proxy      | String  | Proxy |
| timeout    | Float64 | Dial timeout |
| verifycert | Bool    | Whether verify server cert |
| host       | String  | Set a host different with url |

**Returns**

[http.response](#httpresponse) or (nil, error message)

### http.patch(url [, options])

**Attributes**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| url     | String | URL of the resource to load |
| options | Table  | Additional options |

**Options**

| Name       | Type    | Description |
| ---------- | ------- | ----------- |
| params     | Table   | URL encoded query params |
| headers    | Table   | Additional headers to send with the request |
| cookies    | Table   | Additional cookies to send with the request |
| data       | Table   | Deprecated. URL encoded request body. This will also set the `Content-Type` header to `application/x-www-form-urlencoded` |
| rawdata    | String  | Raw request body. |
| json       | String  | Json, This will also set the `Content-Type` header to `application/json` |
| files      | Table   | Upload files, example {file="filepath/file.txt"}, This will also set the `Content-Type` header to `multipart/form-data`, at the same time, you can add data params |
| proxy      | String  | Proxy |
| timeout    | Float64 | Dial timeout |
| redirect   | Bool    | Whether follow redirect |
| verifycert | Bool    | Whether verify server cert |
| host       | String  | Set a host different with url |


**Returns**

[http.response](#httpresponse) or (nil, error message)

### http.post(url [, options])

**Attributes**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| url     | String | URL of the resource to load |
| options | Table  | Additional options |

**Options**

| Name       | Type    | Description |
| ---------- | ------- | ----------- |
| params     | Table   | URL encoded query params |
| headers    | Table   | Additional headers to send with the request |
| cookies    | Table   | Additional cookies to send with the request |
| data       | Table   | Deprecated. URL encoded request body. This will also set the `Content-Type` header to `application/x-www-form-urlencoded` |
| rawdata    | String  | Raw request body. |
| json       | String  | Json, This will also set the `Content-Type` header to `application/json` |
| files      | Table   | Upload files, example {file="filepath/file.txt"}, This will also set the `Content-Type` header to `multipart/form-data`, at the same time, you can add data params |
| proxy      | String  | Proxy |
| timeout    | Float64 | Dial timeout |
| redirect   | Bool    | Whether follow redirect |
| verifycert | Bool    | Whether verify server cert |
| host       | String  | Set a host different with url |

**Returns**

[http.response](#httpresponse) or (nil, error message)

### http.put(url [, options])

**Attributes**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| url     | String | URL of the resource to load |
| options | Table  | Additional options |

**Options**

| Name       | Type    | Description |
| ---------- | ------- | ----------- |
| params     | Table   | URL encoded query params |
| headers    | Table   | Additional headers to send with the request |
| cookies    | Table   | Additional cookies to send with the request |
| data       | Table   | Deprecated. URL encoded request body. This will also set the `Content-Type` header to `application/x-www-form-urlencoded` |
| rawdata    | String  | Raw request body. |
| json       | String  | Json, This will also set the `Content-Type` header to `application/json` |
| files      | Table   | Upload files, example {file="filepath/file.txt"}, This will also set the `Content-Type` header to `multipart/form-data`, at the same time, you can add data params |
| proxy      | String  | Proxy |
| timeout    | Float64 | Dial timeout |
| redirect   | Bool    | Whether follow redirect |
| verifycert | Bool    | Whether verify server cert |
| host       | String  | Set a host different with url |

**Returns**

[http.response](#httpresponse) or (nil, error message)

### http.options(url [, options])

**Attributes**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| url     | String | URL of the resource to load |
| options | Table  | Additional options |

**Options**

| Name       | Type    | Description |
| ---------- | ------- | ----------- |
| params     | Table   | URL encoded query params |
| cookies    | Table   | Additional cookies to send with the request |
| headers    | Table   | Additional headers to send with the request |
| proxy      | String  | Proxy |
| timeout    | Float64 | Dial timeout |
| verifycert | Bool    | Whether verify server cert |
| host       | String  | Set a host different with url |

**Returns**

[http.response](#httpresponse) or (nil, error message)


### http.response

The `http.response` table contains information about a completed HTTP request.

**Attributes**

| Name        | Type   | Description |
| ----------- | ------ | ----------- |
| body        | String | The HTTP response body |
| body_size   | Number | The size of the HTTP reponse body in bytes |
| header      | Table  | The HTTP response headers |
| raw_header  | String | The HTTP response raw headers |
| cookie      | Table  | The cookies sent by the server in the HTTP response |
| raw_cookie  | String | The formated cookies sent by the server in the HTTP response |
| status_code | Number | The HTTP response status code |
| url         | String | The final URL the request ended pointing to after redirects |
| req_scheme  | String | The scheme of request |
| raw_req     | String | The raw request |
| proto       | String | The proto of request |