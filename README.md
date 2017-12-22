# gluahttp

gluahttp provides an easy way to make HTTP requests from within [GopherLua](https://github.com/yuin/gopher-lua).

## Installation

```
go get github.com/Greyh4t/gluahttp
```

## Example

```go
package main

import (
	"time"

	"github.com/Greyh4t/dnscache"
	"github.com/Greyh4t/gluahttp"
	"github.com/yuin/gopher-lua"
)

var jsonStr = `
local json = {}

-- Internal functions.

local function kind_of(obj)
  if type(obj) ~= 'table' then return type(obj) end
  local i = 1
  for _ in pairs(obj) do
    if obj[i] ~= nil then i = i + 1 else return 'table' end
  end
  if i == 1 then return 'table' else return 'array' end
end

local function escape_str(s)
  local in_char  = {'\\', '"', '/', '\b', '\f', '\n', '\r', '\t'}
  local out_char = {'\\', '"', '/',  'b',  'f',  'n',  'r',  't'}
  for i, c in ipairs(in_char) do
    s = s:gsub(c, '\\' .. out_char[i])
  end
  return s
end

-- Returns pos, did_find; there are two cases:
-- 1. Delimiter found: pos = pos after leading space + delim; did_find = true.
-- 2. Delimiter not found: pos = pos after leading space;     did_find = false.
-- This throws an error if err_if_missing is true and the delim is not found.
local function skip_delim(str, pos, delim, err_if_missing)
  pos = pos + #str:match('^%s*', pos)
  if str:sub(pos, pos) ~= delim then
    if err_if_missing then
      error('Expected ' .. delim .. ' near position ' .. pos)
    end
    return pos, false
  end
  return pos + 1, true
end

-- Expects the given pos to be the first character after the opening quote.
-- Returns val, pos; the returned pos is after the closing quote character.
local function parse_str_val(str, pos, val)
  val = val or ''
  local early_end_error = 'End of input found while parsing string.'
  if pos > #str then error(early_end_error) end
  local c = str:sub(pos, pos)
  if c == '"'  then return val, pos + 1 end
  if c ~= '\\' then return parse_str_val(str, pos + 1, val .. c) end
  -- We must have a \ character.
  local esc_map = {b = '\b', f = '\f', n = '\n', r = '\r', t = '\t'}
  local nextc = str:sub(pos + 1, pos + 1)
  if not nextc then error(early_end_error) end
  return parse_str_val(str, pos + 2, val .. (esc_map[nextc] or nextc))
end

-- Returns val, pos; the returned pos is after the number's final character.
local function parse_num_val(str, pos)
  local num_str = str:match('^-?%d+%.?%d*[eE]?[+-]?%d*', pos)
  local val = tonumber(num_str)
  if not val then error('Error parsing number at position ' .. pos .. '.') end
  return val, pos + #num_str
end


-- Public values and functions.

function json.encode(obj, as_key)
  local s = {}  -- We'll build the string as an array of strings to be concatenated.
  local kind = kind_of(obj)  -- This is 'array' if it's an array or type(obj) otherwise.
  if kind == 'array' then
    if as_key then error('Can\'t encode array as key.') end
    s[#s + 1] = '['
    for i, val in ipairs(obj) do
      if i > 1 then s[#s + 1] = ', ' end
      s[#s + 1] = json.encode(val)
    end
    s[#s + 1] = ']'
  elseif kind == 'table' then
    if as_key then error('Can\'t encode table as key.') end
    s[#s + 1] = '{'
    for k, v in pairs(obj) do
      if #s > 1 then s[#s + 1] = ', ' end
      s[#s + 1] = json.encode(k, true)
      s[#s + 1] = ':'
      s[#s + 1] = json.encode(v)
    end
    s[#s + 1] = '}'
  elseif kind == 'string' then
    return '"' .. escape_str(obj) .. '"'
  elseif kind == 'number' then
    if as_key then return '"' .. tostring(obj) .. '"' end
    return tostring(obj)
  elseif kind == 'boolean' then
    return tostring(obj)
  elseif kind == 'nil' then
    return 'null'
  else
    error('Unjsonifiable type: ' .. kind .. '.')
  end
  return table.concat(s)
end

json.null = {}  -- This is a one-off table to represent the null value.

function json.decode(str, pos, end_delim)
  pos = pos or 1
  if pos > #str then error('Reached unexpected end of input.') end
  local pos = pos + #str:match('^%s*', pos)  -- Skip whitespace.
  local first = str:sub(pos, pos)
  if first == '{' then  -- Parse an object.
    local obj, key, delim_found = {}, true, true
    pos = pos + 1
    while true do
      key, pos = json.decode(str, pos, '}')
      if key == nil then return obj, pos end
      if not delim_found then error('Comma missing between object items.') end
      pos = skip_delim(str, pos, ':', true)  -- true -> error if missing.
      obj[key], pos = json.decode(str, pos)
      pos, delim_found = skip_delim(str, pos, ',')
    end
  elseif first == '[' then  -- Parse an array.
    local arr, val, delim_found = {}, true, true
    pos = pos + 1
    while true do
      val, pos = json.decode(str, pos, ']')
      if val == nil then return arr, pos end
      if not delim_found then error('Comma missing between array items.') end
      arr[#arr + 1] = val
      pos, delim_found = skip_delim(str, pos, ',')
    end
  elseif first == '"' then  -- Parse a string.
    return parse_str_val(str, pos + 1)
  elseif first == '-' or first:match('%d') then  -- Parse a number.
    return parse_num_val(str, pos)
  elseif first == end_delim then  -- End of an object or array.
    return nil, pos + 1
  else  -- Parse true, false, or null.
    local literals = {['true'] = true, ['false'] = false, ['null'] = json.null}
    for lit_str, lit_val in pairs(literals) do
      local lit_end = pos + #lit_str - 1
      if str:sub(pos, lit_end) == lit_str then return lit_val, lit_end + 1 end
    end
    local pos_info_str = 'position ' .. pos .. ': ' .. str:sub(pos, pos + 10)
    error('Invalid json syntax starting at ' .. pos_info_str)
  end
end

return json
`

func main() {
	L := lua.NewState()
	defer L.Close()

	L.DoString(jsonStr)
	j := L.Get(-1)
	L.Pop(1)
	loaded := L.GetField(L.Get(-10000), "_LOADED")
	L.SetField(loaded, "json", j)

	// resolver为dns缓存
	// 如果不需要使用dns缓存，New(nil)即可
	resolver := dnscache.New(time.Minute * 10)
	L.PreloadModule("http", gluahttp.New(resolver).Loader)

	if err := L.DoString(`
local json = require("json")
local http = require("http")

-- 支持get, post, head, delete, patch, put, options
resp, err = http.get("http://passport.jd.com/new/login.aspx?ReturnUrl=http%3A%2F%2Fhome.jd.com%2F", { 
	-- 请求超时，默认30秒
	timeout = 10,

	-- 是否校验服务端证书，默认true
	-- verify = false,

	-- 是否允许重定向，默认true
	-- redirect = false,

	-- 是否允许请求gzip格式，默认true
	-- compress = false,

	-- 是否添加ajax头，默认false
	-- ajax = true,

	-- 自定义host头
	-- host = "www.jd.com"

	-- 401基础认证
	-- auth = {"username","password"},

	-- http请求头
	headers = {
		Test="xxx",
		["User-Agent"]="testUserAgent",
	},

	-- 自定义cookie
	cookies = {
		session="xxx",
		user="test"
	},

	-- 代理
	proxies = {
		http="http://127.0.0.1:8080",
		https="http://127.0.0.1:8080"
	},

	-- query参数，参数会被url编码
	-- params = {
	-- 	"q1"="测试",
	-- 	"q2"="bbb"
	-- },

	-- 原始query，不被url编码，如果设置了raw_query，params会被忽略
	-- raw_query = "q1=测试&q2=bbb",

	-- 请求body，参数会被url编码
	-- data = {
	-- 	data1="测试",
	-- 	data2="aaa"
	-- },

	-- 原始请求body，不被url编码，需要自行设置Content-Type头，如果设置了raw_data，data、json、xml、files参数将被忽略
	-- raw_data = "data1=测试&data2=aaa",

	-- 上传文件，key为字段名，value为文件名，post方法可同时上传多个
	-- files={
	-- 	file="test.txt"
	-- },

	-- 发送json，参数为json格式字符串
	-- json = '{"a":"d","b":[{"c":1},{"d":2}]}',

	-- 发送xml，参数为xml格式字符串
	-- xml = '<?xml version="1.0" encoding="ISO-8859-1"?>' ..
	-- 		'<!--  Copyright w3school.com.cn -->' ..
	-- 		'<note>' ..
	-- 		'<to>George</to>' ..
	-- 		'<from>John</from>' ..
	-- 		'<heading>Reminder</heading>' ..
	-- 		"<body>Don't forget the meeting!</body>" ..
	-- 	'</note>',
})
if err then
	print(err)
	return
end
print(json.encode(resp))
	`); err != nil {
		panic(err)
	}
}
```

## Response example

```
{
    "status_code": 200,
    "body": "<!DOCTYPE html>\r\n<html>\r\n<head>\r\n    <meta charset=\"UTF-8\"\/>\r\n    <meta http-equiv=\"X-UA-Compatible\" content=\"IE=Edge\"\/>\r\n    <title>京东<\/title>\r\n<\/body>\r\n<\/html>\r\n",
    "body_size": 15579,
    "headers": {
        "Pragma": "no-cache,",
        "Server": "jfe,",
        "Cache-Control": "max-age=0,",
        "Expires": "Mon, 18 Dec 2017 09:19:54 GMT,",
        "Set-Cookie": "qr_t=c; Path=\/; HttpOnly,alc=CMDyrO3bMtUxD6DofCkq+w==; Path=\/; HttpOnly;,_t=wR2FB6ybw8RML13oirnDVqNsenKXTcxQGy\/tisG3EDE=; Path=\/;,",
        "Content-Language": "zh-CN,",
        "Date": "Mon, 18 Dec 2017 09:19:54 GMT,",
        "Content-Length": "15579,",
        "Content-Type": "text\/html;charset=GBK,",
        "Vary": "Accept-Encoding,"
    },
    "raw_headers": "Content-Type: text\/html;charset=GBK\r\nVary: Accept-Encoding\r\nPragma: no-cache\r\nSet-Cookie: qr_t=c; Path=\/; HttpOnly\r\nSet-Cookie: alc=CMDyrO3bMtUxD6DofCkq+w==; Path=\/; HttpOnly;\r\nSet-Cookie: _t=wR2FB6ybw8RML13oirnDVqNsenKXTcxQGy\/tisG3EDE=; Path=\/;\r\nContent-Language: zh-CN\r\nServer: jfe\r\nCache-Control: max-age=0\r\nExpires: Mon, 18 Dec 2017 09:19:54 GMT\r\nDate: Mon, 18 Dec 2017 09:19:54 GMT\r\nContent-Length: 15579",
    "cookies": {
        "qr_t": "c",
        "alc": "CMDyrO3bMtUxD6DofCkq+w==",
        "_t": "wR2FB6ybw8RML13oirnDVqNsenKXTcxQGy\/tisG3EDE="
    },
    "raw_cookies": "qr_t=c;alc=CMDyrO3bMtUxD6DofCkq+w==;_t=wR2FB6ybw8RML13oirnDVqNsenKXTcxQGy\/tisG3EDE=",
    "proto": "HTTP\/1.1",
    "url": "https:\/\/passport.jd.com\/new\/login.aspx?ReturnUrl=http%3A%2F%2Fhome.jd.com%2F",
    "request": {
        "method": "GET",
        "url": "https:\/\/passport.jd.com\/new\/login.aspx?ReturnUrl=http%3A%2F%2Fhome.jd.com%2F",
        "scheme": "https",
        "proto": "",
        "host": "passport.jd.com",
        "body": "",
        "headers": {
            "X-Scanner": "ZERO,",
            "Test": "xxx,",
            "User-Agent": "testUserAgent,",
            "Cookie": "session=xxx; user=test,",
            "Referer": "http:\/\/passport.jd.com\/new\/login.aspx?ReturnUrl=http%3A%2F%2Fhome.jd.com%2F,"
        },
        "raw_headers": "X-Scanner: ZERO\r\nTest: xxx\r\nUser-Agent: testUserAgent\r\nCookie: session=xxx; user=test\r\nReferer: http:\/\/passport.jd.com\/new\/login.aspx?ReturnUrl=http%3A%2F%2Fhome.jd.com%2F",
        "cookies": {
            "session": "xxx",
            "user": "test"
        },
        "raw_cookies": "session=xxx;user=test",
        "raw": "GET \/new\/login.aspx?ReturnUrl=http%3A%2F%2Fhome.jd.com%2F \r\nHost: passport.jd.com\r\nCookie: session=xxx; user=test\r\nReferer: http:\/\/passport.jd.com\/new\/login.aspx?ReturnUrl=http%3A%2F%2Fhome.jd.com%2F\r\nX-Scanner: ZERO\r\nTest: xxx\r\nUser-Agent: testUserAgent\r\n\r\n"
    },
    "history": [{
            "status_code": 301,
            "body": "",
            "body_size": 0,
            "headers": {
                "Date": "Mon, 18 Dec 2017 09:19:53 GMT,",
                "Content-Type": "text\/html,",
                "Content-Length": "178,",
                "Location": "https:\/\/passport.jd.com\/new\/login.aspx?ReturnUrl=http%3A%2F%2Fhome.jd.com%2F,",
                "Server": "jfe,"
            },
            "raw_headers": "Content-Length: 178\r\nLocation: https:\/\/passport.jd.com\/new\/login.aspx?ReturnUrl=http%3A%2F%2Fhome.jd.com%2F\r\nServer: jfe\r\nDate: Mon, 18 Dec 2017 09:19:53 GMT\r\nContent-Type: text\/html",
            "cookies": {},
            "raw_cookies": "",
            "proto": "HTTP\/1.1",
            "url": "http:\/\/passport.jd.com\/new\/login.aspx?ReturnUrl=http%3A%2F%2Fhome.jd.com%2F",
            "request": {
                "method": "GET",
                "url": "http:\/\/passport.jd.com\/new\/login.aspx?ReturnUrl=http%3A%2F%2Fhome.jd.com%2F",
                "scheme": "http",
                "proto": "HTTP\/1.1",
                "host": "passport.jd.com",
                "body": "",
                "headers": {
                    "X-Scanner": "ZERO,",
                    "Test": "xxx,",
                    "User-Agent": "testUserAgent,",
                    "Cookie": "session=xxx; user=test,"
                },
                "raw_headers": "Test: xxx\r\nUser-Agent: testUserAgent\r\nCookie: session=xxx; user=test\r\nX-Scanner: ZERO",
                "cookies": {
                    "session": "xxx",
                    "user": "test"
                },
                "raw_cookies": "session=xxx;user=test",
                "raw": "GET \/new\/login.aspx?ReturnUrl=http%3A%2F%2Fhome.jd.com%2F HTTP\/1.1\r\nHost: passport.jd.com\r\nX-Scanner: ZERO\r\nTest: xxx\r\nUser-Agent: testUserAgent\r\nCookie: session=xxx; user=test\r\n\r\n"
            }
        }
    ]
}
```

## 参考

[grequests](https://github.com/levigross/grequests)
[gluahttp](https://github.com/cjoudrey/gluahttp)
