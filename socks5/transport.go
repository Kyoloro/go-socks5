package socks5

import (
	"log"
	"net"
)

type (
	Transport interface {
		Pip()
	}

	secureTransport struct {
		remote net.Conn
		client net.Conn
	}
)

func (st *secureTransport) Pip() {
	defer func() {
		if p := recover(); p != nil {
			log.Println(p)
		}
	}()

	defer st.remote.Close()

	go func() {
		for {
			b, err := st.readFromRemote()

			if err != nil {
				break
			}

			log.Printf("Remote -> Client %dbyte\n", len(b))
			_, err = st.writeToClient(b)

			if err != nil {
				break
			}
		}
	}()

	for {
		b, err := st.readFromClient()

		if err != nil {
			break
		}

		log.Printf("Client -> Remote %dbyte\n", len(b))
		_, err = st.writeToRemote(b)

		if err != nil {
			break
		}
	}
}

func (st *secureTransport) readFromRemote() ([]byte, error) {
	b := make([]byte, ServerConfig.TcpBufferSize)
	n, err := st.remote.Read(b)

	if err != nil {
		log.Println(err)
	}

	return b[:n], err
}

func (st *secureTransport) writeToRemote(b []byte) (int, error) {
	n, err := st.remote.Write(b)

	if err != nil {
		log.Println(err)
	}

	return n, err
}

func (st *secureTransport) readFromClient() ([]byte, error) {
	b := make([]byte, ServerConfig.TcpBufferSize)
	n, err := st.client.Read(b)

	if err != nil {
		log.Println(err)
	}

	return b[:n], err
}

func (st *secureTransport) writeToClient(b []byte) (int, error) {
	n, err := st.client.Write(b)

	if err != nil {
		log.Println(err)
	}

	return n, err
}

func NewSecureTransport(remote, client net.Conn) Transport {
	st := &secureTransport{
		remote: remote,
		client: client,
	}

	return st
}
