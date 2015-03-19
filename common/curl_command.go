package common

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

type DataType int

const (
	DataAsciiType DataType = iota
	DataBinaryType
	DataUrlEncodeType
	FormType
	FormStringType
)

type DataOption struct {
	Value string
	Type  DataType
}

func (self *DataOption) IsFormStyle() bool {
	return strings.IndexByte(self.Value, '=') != -1
}

func (self *DataOption) UseExternalFile() bool {
	return self.FileName() != ""
}

func (self *DataOption) FileName() string {
	if self.Type == FormType {
		index := strings.Index(self.Value, "=")
		if index == -1 {
			return ""
		}
		if index < len(self.Value)-1 {
			nextChar := self.Value[index+1 : index+2]
			if nextChar == "@" || nextChar == "<" {
				return strings.Split(self.Value[index+2:], ";")[0]
			}
		}
	} else if self.Type != FormStringType {
		if strings.HasPrefix(self.Value, "@") {
			return strings.Split(self.Value[1:], ";")[0]
		}
	}
	return ""
}

func (self *DataOption) SendAsFormFile() bool {
	if self.Type == FormType {
		index := strings.Index(self.Value, "=")
		if index == -1 {
			return false
		}
		if index < len(self.Value)-1 {
			nextChar := self.Value[index+1 : index+2]
			if nextChar == "@" {
				return true
			}
		}
	}
	return false
}

type DataOptions []DataOption

func (self *DataOptions) Append(data string, typeEmum DataType) {
	*self = append(*self, DataOption{Value: data, Type: typeEmum})
}

func (self *DataOptions) HasAnyData() bool {
	return len(*self) > 0
}

func (self *DataOptions) HasData() bool {
	for _, data := range *self {
		switch data.Type {
		case DataAsciiType:
			return true
		case DataBinaryType:
			return true
		case DataUrlEncodeType:
			return true
		}
	}
	return false
}

func (self *DataOptions) HasForm() bool {
	for _, data := range *self {
		switch data.Type {
		case FormType:
			return true
		case FormStringType:
			return true
		}
	}
	return false
}

func (self *DataOptions) ExternalFileCount() int {
	count := 0
	for _, data := range *self {
		if data.UseExternalFile() {
			count++
		}
	}
	return count
}

type CurlOptions struct {
	// Example of verbosity with level
	Basic         bool         `long:"basic" description:"Use HTTP Basic Authentication (H)"`
	Compressed    func()       `long:"compressed" description:"Request compressed response (using deflate or gzip)"`
	Cookie        []string     `short:"b" long:"cookie" value-name:"STRING/FILE" description:"Read cookies from STRING/FILE (H)"`
	CookieJar     string       `short:"c" long:"cookie-jar" value-name:"FILE" description:"Write cookies to FILE after operation (H)"`
	Data          func(string) `short:"d" long:"data" value-name:"DATA" description:"HTTP POST data (H)"`
	DataAscii     func(string) `long:"data-ascii" value-name:"DATA" description:"HTTP POST ASCII data (H)"`
	DataBinary    func(string) `long:"data-binary" value-name:"DATA" description:"HTTP POST binary data (H)"`
	DataUrlEncode func(string) `long:"data-urlencode" value-name:"DATA" description:"HTTP POST data url encoded (H)"`
	//Digest bool `long:"digest" description:"Use HTTP Digest Authentication (H)"`
	Get        bool         `short:"G" long:"get" description:"Send the -d data with a HTTP GET (H)"`
	Form       func(string) `short:"F" long:"form" value-name:"KEY=VALUE" description:"Specify HTTP multipart POST data (H)"`
	FormString func(string) `long:"form-string" value-name:"KEY=VALUE" description:"Specify HTTP multipart POST data (H)"`
	Header     []string     `short:"H" long:"header" value-name:"LINE" description:"Pass custom header LINE to server (H)"`
	Head       bool         `short:"I" long:"head" description:"Show document info only"`
	//Http11 bool `long:"http1.1" description:"Use HTTP 1.1 (H)"`
	//Http2 bool `long:"http2" description:"Use HTTP 2 (H)"`
	Proxy      string       `short:"x" long:"proxy" value-name:"[PROTOCOL://]HOST[:PORT]" description:"Use proxy on given port"`
	Referer    func(string) `short:"e" long:"referer" description:"Referer URL (H)"`
	Request    string       `short:"X" long:"request" value-name:"COMMAND" description:"Specify request command to use"`
	TrEncoding func()       `long:"tr-encoding" description:"Request compressed transfer encoding (H)"`
	Transfer   func(string) `short:"T" long:"upload-file" value-name:"FILE" description:"Transfer FILE to destination"`
	Url        string       `long:"url" value-name:"URL" description:"URL to work with"`
	User       string       `short:"u" long:"user" value-name:"USER[:PASSWORD]" description:"Server user and password"`
	UserAgent  func(string) `short:"A" long:"user-agent" value-name:"STRING" description:"User-Agent to send to server (H)"`

	// Original parameter
	AWSV2 string `long:"awsv2" value-name:"ACCESS-KEY:SECRET-KEY" description:"AWS V2 style authentication (original)"`

	// Internal Use
	ProcessedData DataOptions
}

func (self *CurlOptions) Init() {
	self.Compressed = func() {
		self.Header = append(self.Header, "Accept-Encoding: deflate", "Accept-Encoding: gzip")
	}

	self.Data = func(data string) {
		self.ProcessedData.Append(data, DataAsciiType)
	}
	self.DataAscii = self.Data

	self.DataBinary = func(data string) {
		self.ProcessedData.Append(data, DataBinaryType)
	}

	self.DataUrlEncode = func(data string) {
		self.ProcessedData.Append(data, DataUrlEncodeType)
	}

	self.Form = func(data string) {
		self.ProcessedData.Append(data, FormType)
	}

	self.FormString = func(data string) {
		self.ProcessedData.Append(data, FormStringType)
	}

	self.Referer = func(data string) {
		self.Header = append(self.Header, fmt.Sprintf("Referer: %s", data))
	}

	self.Transfer = func(data string) {
		self.ProcessedData.Append(fmt.Sprintf("@%s", data), DataBinaryType)
		if self.Request == "" {
			self.Request = "PUT"
		}
	}

	self.TrEncoding = func() {
		self.Header = append(self.Header, "Te: gzip")
	}

	self.UserAgent = func(data string) {
		self.Header = append(self.Header, fmt.Sprintf("User-Agent: %s", data))
	}

}

func (self *CurlOptions) CheckError() error {
	if self.ProcessedData.HasData() && self.ProcessedData.HasData() {
		return fmt.Errorf("Warning: You can only select one HTTP request!")
	}
	return nil
}

func (self *CurlOptions) Method() string {
	method := strings.ToUpper(self.Request)
	// explicit method is the highest priority
	if method != "" {
		return method
	}
	if self.Get {
		return "GET"
	}
	if self.Head {
		return "HEAD"
	}
	if self.ProcessedData.HasAnyData() {
		return "POST"
	}
	return "GET"
}

func (self *CurlOptions) Headers() [][]string {
	var result [][]string
	for _, header := range self.Header {
		words := strings.SplitN(header, ":", 2)
		if len(words) != 2 {
			fmt.Fprintf(os.Stderr, "[warning] %s is wrong style header.\n", header)
			continue
		}
		words[1] = strings.TrimSpace(words[1])
		result = append(result, words)
	}
	return result
}

type HeaderGroup struct {
	Key    string
	Values []string
}

func (self *CurlOptions) GroupedHeaders() []HeaderGroup {
	headers := self.Headers()
	index := make(map[string]int)
	var result []HeaderGroup

	for _, header := range headers {
		key := strings.ToLower(header[0])
		i, ok := index[key]
		if ok {
			result[i].Values = append(result[i].Values, header[1])
		} else {
			headerGroup := HeaderGroup{Key: key, Values: make([]string, 1)}
			headerGroup.Values[0] = header[1]
			index[key] = len(result)
			result = append(result, headerGroup)
		}
	}
	return result
}

func (self *CurlOptions) OnlyHasContentTypeHeader() bool {
	headers := self.Headers()
	if len(headers) == 0 || (len(headers) == 1 && strings.ToLower(headers[0][0]) == "content-type") {
		return true
	}
	return false
}

func (self *CurlOptions) FindContentTypeHeader() string {
	headers := self.Header
	for _, header := range headers {
		fragments := strings.SplitN(header, ":", 2)
		if len(fragments) == 2 && strings.TrimSpace(strings.ToLower(fragments[0])) == "content-type" {
			return strings.TrimSpace(fragments[1])
		}
	}
	return ""
}

func (self *CurlOptions) InsertContentTypeHeader(contentType string) {
	contentTypeInHeader := self.FindContentTypeHeader()
	if contentTypeInHeader == "" {
		self.Header = append(self.Header, fmt.Sprintf("Content-Type: %s", contentType))
	}
}

func (self *CurlOptions) CanUseSimpleForm() bool {
	for _, data := range self.ProcessedData {
		if data.UseExternalFile() {
			return false
		}
		singleData, err := url.ParseQuery(data.Value)
		if len(singleData) == 0 || err != nil {
			return false
		}
		for _, values := range singleData {
			if len(values) == 1 && values[0] == "" {
				return false
			}
		}
	}
	return true
}
