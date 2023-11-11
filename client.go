package golokia

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Target struct {
	URL      string `json:"url"`
	Username string `json:"user"`
	Password string `json:"password"`
}

const (
	ReadRequest   = "read"
	ListRequest   = "list"
	SearchRequest = "search"
	ExecRequest   = "exec"
)

type Request struct {
	Type string `json:"type"`
	// MBean's ObjectName which can be a pattern
	Mbean string `json:"mbean,omitempty"`
	// Inner path for accessing the value of a complex value
	Path string `json:"path,omitempty"`
	// Attribute name to read or a JSON array containing a list of attributes to read. No attribute is given, then all attributes are read.
	Attribute string `json:"attribute,omitempty"`

	// The operation to execute, optionally with a signature as described above. 	dumpAllThreads
	Operation string `json:"operation"`

	// An array of arguments for invoking this operation. The value must be serializable as described in Section 6.4.2, “Request parameter serialization”. 	[true,true]
	Arguments []interface{} `json:"arguments"`

	Target *Target `json:"target,omitempty"`

	Config *Options `json:"config,omitempty"`
}

type Options struct {
	// Maximum depth of the tree traversal into a bean's properties. The maximum value as configured in the agent's configuration is a hard limit and cannot be exceeded by a query parameter.
	MaxDepth int `json:"maxDepth,omitempty"`

	// For collections (lists, maps) this is the maximum size.
	MaxCollectionSize int `json:"maxCollectionSize,omitempty"`

	// Number of objects to visit in total. A hard limit can be configured in the agent's configuration.
	MaxObjects int `json:"maxObjects,omitempty"`

	// If set to "true", a Jolokia operation will not return an error if an JMX operation fails, but includes the exception message as value. This is useful for e.g. the read operation when requesting multiple attributes' values. Default: false
	IgnoreErrors bool `json:"ignoreErrors,omitempty"`

	// The MIME type to return for the response. By default, this is text/plain, but it can be useful for some tools to change it to application/json. Init parameters can be used to change the default mime type. Only text/plain and application/json are allowed. For any other value Jolokia will fallback to text/plain.
	MimeType string `json:"mimeType,omitempty"`

	// Defaults to true to return the canonical format of property lists. If set to false then the default unsorted property list is returned.
	CanonicalNaming bool `json:"canonicalNaming,omitempty"`

	// If set to true, then in case of an error the stack trace is included. With false no stack trace will be returned, and when this parameter is set to runtime only for RuntimeExceptions a stack trace is put into the error response. Default is true if not set otherwise in the global agent configuration.
	IncludeStackTrace bool `json:"includeStackTrace,omitempty"`

	// If this parameter is set to true then a serialized version of the exception is included in an error response. This value is put under the key error_value in the response value. By default this is set to false except when the agent global configuration option is configured otherwise.
	SerializeException bool `json:"serializeException,omitempty"`

	// If this parameter is given, its value is interpreted as epoch time (seconds since 1.1.1970) and if the requested value did not change since this time, an empty response (with no value) is returned and the response status code is set to 304 ("Not modified"). This option is currently only supported for LIST requests. The time value can be extracted from a previous' response timestamp.
	IfModifiedSince bool `json:"ifModifiedSince,omitempty"`

	responseBody interface{} `json:"-"`
}

type Response struct {
	Timestamp int64                  `json:"timestamp"`
	Status    int64                  `json:"status"`
	Request   map[string]interface{} `json:"request,omitempty"`
	Value     interface{}            `json:"value,omitempty"`
}

type Client struct {
	BaseURL  string `json:"base_url"`
	Username string `json:"username"`
	Password string `json:"password"`

	Client *http.Client
}

func (client *Client) Do(ctx context.Context, r *Request, responseValue interface{}) (*Response, error) {
	body, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(body))

	request, err := http.NewRequestWithContext(ctx, "POST", client.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	if client.Username != "" {
		request.SetBasicAuth(client.Username, client.Password)
	}

	var response *http.Response
	if client.Client == nil {
		response, err = http.DefaultClient.Do(request)
	} else {
		response, err = client.Client.Do(request)
	}
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("status code: %d", response.StatusCode))
	}

	var responseBody Response
	responseBody.Value = responseValue
	decoder := json.NewDecoder(response.Body)
	decoder.UseNumber()
	err = decoder.Decode(&responseBody)
	if err != nil {
		return nil, err
	}
	return &responseBody, nil
}

func toConfig(opts []*Options) *Options {
	if len(opts) > 0 {
		return opts[0]
	}
	return nil
}

func toResponseBody(opts []*Options) interface{} {
	if len(opts) > 0 {
		if opts[0] == nil {
			return nil
		}
		return opts[0].responseBody
	}
	return nil
}

func (client *Client) Read(ctx context.Context, target *Target, mbean, attribute, path string, opts ...*Options) (*Response, error) {
	return client.Do(ctx, &Request{
		Type:      ReadRequest,
		Mbean:     mbean,
		Attribute: attribute,
		Path:      path,
		Target:    target,
		Config:    toConfig(opts),
	},
		toResponseBody(opts))
}

func (client *Client) Exec(ctx context.Context, target *Target, mbean, operation string, arguments []interface{}, opts ...*Options) (*Response, error) {
	return client.Do(ctx, &Request{
		Type:      ExecRequest,
		Mbean:     mbean,
		Operation: operation,
		Arguments: arguments,
		Target:    target,
		Config:    toConfig(opts),
	},
		toResponseBody(opts))
}

func (client *Client) Search(ctx context.Context, target *Target, mbean string, opts ...*Options) (*Response, error) {
	return client.Do(ctx, &Request{
		Type:      SearchRequest,
		Mbean:     mbean,
		Target:    target,
		Config:    toConfig(opts),
	},
		toResponseBody(opts))
}

func (client *Client) List(ctx context.Context, target *Target, path string, opts ...*Options) (*Response, error) {
	return client.Do(ctx, &Request{
		Type:   ListRequest,
		Path:   path,
		Target: target,
		Config: toConfig(opts),
	},
		toResponseBody(opts))
}

func (client *Client) ListDomains(ctx context.Context, target *Target, opts ...*Options) ([]string, error) {
	opt := toConfig(opts)
	if opt == nil {
		opt = &Options{}
	}
	opt.MaxDepth = 1
	response, err := client.List(ctx, target, "", opt)
	if err != nil {
		return nil, err
	}
	return extractKeys(response.Value)
}

func (client *Client) ListBeans(ctx context.Context, target *Target, domain string, opts ...*Options) ([]string, error) {
	opt := toConfig(opts)
	if opt == nil {
		opt = &Options{}
	}
	opt.MaxDepth = 1
	response, err := client.List(ctx, target, domain, opt)
	if err != nil {
		return nil, err
	}
	return extractKeys(response.Value)
}

func extractKeys(o interface{}) ([]string, error) {
	m, ok := o.(map[string]interface{})
	if !ok {
		return nil, errors.New("type of response value isn't match")
	}
	var names = make([]string, 0, len(m))
	for k, _ := range m {
		names = append(names, k)
	}
	return names, nil
}

type Attr struct {
	Rw   bool   `json:"rw"`
	Type string `json:"type"`
	Desc string `json:"desc"`
}

type Op struct {
	Args []interface{} `json:"args"`
	Ret  string        `json:"ret"`
	Desc string        `json:"desc"`
}

type MbeanClass struct {
	OpList map[string]Op   `json:"op"`
	Attrs  map[string]Attr `json:"attr"`
	Class  string          `json:"class"`
	Desc   string          `json:"desc"`
}

func (client *Client) ReadClass(ctx context.Context, target *Target, domain, attributes string, opts ...*Options) (*MbeanClass, error) {
	var opt *Options
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	} else {
		opt = &Options{}
	}

	opt.responseBody = &MbeanClass{}

	response, err := client.List(ctx, target, domain+":"+attributes, toConfig(opts))
	if err != nil {
		return nil, err
	}
	cls, ok := response.Value.(*MbeanClass)
	if !ok {
		return nil, fmt.Errorf("result unexpected - %T", response.Value)
	}
	return cls, nil
}

func (client *Client) ListProperties(ctx context.Context, target *Target, mbean, attributes string, opts ...*Options) (map[string]interface{}, *Response, error) {
	var opt *Options
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	} else {
		opt = &Options{}
	}

	var values map[string]interface{}
	opt.responseBody = &values
	response, err := client.Read(ctx, target, mbean, attributes, "", toConfig(opts))
	return values, response, err
}
