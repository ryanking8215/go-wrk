package loader

type RequestConfig struct {
	Method  string            `json:"method"`
	Host    string            `json:"host"`
	Header  map[string]string `json:"header"`
	TestUrl string            `json:"url"`
	ReqBody string            `json:"body"`
}

type Config struct {
	Duration           int    `json:"duration"` //seconds
	Goroutines         int    `json:"goroutines"`
	Timeoutms          int    `json:"timeoutms"`
	AllowRedirects     bool   `json:"redir"`
	DisableCompression bool   `json:"no_comp"`
	DisableKeepAlive   bool   `json:"no_keepalive"`
	SkipVerify         bool   `json:"skip_verify"`
	ClientCert         string `json:"client_cert"`
	ClientKey          string `json:"client_key"`
	CaCert             string `json:"ca_cert"`
	Http2              bool   `json:"http2"`
	RequestConfig
	Id int64 `json:"id"`
}

func NewConfig() Config {
	return Config{
		RequestConfig: RequestConfig{
			Header: make(map[string]string),
		},
	}
}

func (c Config) Clone() Config {
	cloned := c // copy all

	// copy header
	cloned.Header = map[string]string{}
	for k, v := range c.Header {
		cloned.Header[k] = v
	}
	return cloned
}
