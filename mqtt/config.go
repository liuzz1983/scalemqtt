package mqtt

type ServerConfig struct {
	Timeout int
	Address string
}

//LoadConfig load config
func LoadConfig(fileName string) (*ServerConfig, error) {
	return &ServerConfig{
		Timeout: 1,
		Address: ":8080",
	}, nil
}
