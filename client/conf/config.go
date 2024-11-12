package conf

type Server struct {
	Addr string `json:"addr"`
}

type Client struct {
	Addr string   `json:"addr"`
	Cert CertConf `json:"cert"`
}

type Config struct {
	Local *Server `json:"local"`
	Agent *Client `json:"agent"`
}

type CertConf struct {
	Cert string `json:"cert"`
	Key  string `json:"key"`
}
