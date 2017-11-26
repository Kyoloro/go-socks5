package main

import (
	"fmt"
	"go-socks5/socks5"
	"log"
	"net"
)

func main() {
	socks5.InitConfig()

	svr, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", socks5.ServerConfig.Port))

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Socks5 server start port %d", socks5.ServerConfig.Port)

	for {
		conn, err := svr.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("客户端 %s 建立连接\n", conn.RemoteAddr().String())

		sp := socks5.NewServerProtocol(conn)
		go sp.Process()
	}
}
