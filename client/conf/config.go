package conf

type Server struct {
	Addr string `json:"addr"`
}

type Client struct {
	Addr  string   `json:"addr"`
	Key   string   `json:"key"`
	Nonce string   `json:"nonce"`
	Cert  CertConf `json:"cert"`
}

type Config struct {
	Local *Server `json:"local"`
	Agent *Client `json:"agent"`
}

type CertConf struct {
	Cert     string `json:"cert"`
	CertData string `json:"certData"`
	Key      string `json:"key"`
	KeyData  string `json:"keyData"`
	Pfx      string `json:"pfx"`
	Pwd      string `json:"pwd"`
}
