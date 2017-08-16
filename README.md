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

	L.PreloadModule("http", gluahttp.NewHttpModule().Loader)

	if err := L.DoString(`
local http = require("http")
response, error_message = http.post("http://www.jd.com?x=1", {
	params={
		page="测试",
		x="2"
	},
	data={
		a=1,
		b="2"
	},
	--files={
	--	file="/x.txt"
	--},
	--json='{"a":"测试"}',
	--rawdata="aaa'gdgd",
	headers={
		Accept="*/*",
		Test="xxx",
		["User-Agent"]="testUserAgent"
	},
	cookies={
		session="xxx",
		user="test"
	},
	--proxy="http://127.0.0.1:8080",
	timeout=3,
	redirect=false,
	verifycert=false
})
if error_message
then
	print(error_message)
else
	print(response.schema)
	print(response.raw_request)
--	print(response.body)
	print(response.headers["Content-Type"])
	print(response.headers["Location"])
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
| params     | Table   | URL encoded query params |
| cookies    | Table   | Additional cookies to send with the request |
| headers    | Table   | Additional headers to send with the request |
| proxy      | String  | Proxy |
| timeout    | Float64 | Dial timeout |
| redirect   | Bool    | Whether follow redirect |
| verifycert | Bool    | Whether verify server cert |

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
| params     | Table   | URL encoded query params |
| cookies    | Table   | Additional cookies to send with the request |
| headers    | Table   | Additional headers to send with the request |
| proxy      | String  | Proxy |
| timeout    | Float64 | Dial timeout |
| redirect   | Bool    | Whether follow redirect |
| verifycert | Bool    | Whether verify server cert |

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

**Returns**

[http.response](#httpresponse) or (nil, error message)
