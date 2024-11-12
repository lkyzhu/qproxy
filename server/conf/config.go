package conf

type Config struct {
	Addr string   `json:"addr"`
	Cert CertConf `json:"cert"`
}

type CertConf struct {
	Cert string `json:"cert"`
	Key  string `json:"key"`
}
