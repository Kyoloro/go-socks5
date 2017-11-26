package socks5

type (
	serverConfig struct {
		Port          int
		TcpBufferSize int
	}
)

var ServerConfig *serverConfig

func InitConfig() {
	ServerConfig = &serverConfig{
		Port:          1080,
		TcpBufferSize: 2048,
	}
}
