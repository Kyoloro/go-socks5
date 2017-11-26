package socks5

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"

	"go-socks5/util"
)

const (
	Socks5Version = 0x5

	AuthNone   = 0x00
	CmdConnect = 0x01

	AddressTypeIpv4   = 0x01
	AddressTypeDomain = 0x03
	AddressTypeIpv6   = 0x04
)

type (
	processor func([]byte) (processor, error)
	Protocol  interface {
		Process()
	}

	serverProtocol struct {
		conn          net.Conn
		processor     processor
		remoteAddr    string
		remoteAddrRaw []byte
	}
)

func (sp *serverProtocol) Process() {
	for {
		b := make([]byte, ServerConfig.TcpBufferSize)
		n, err := sp.conn.Read(b)
		if err != nil {
			log.Println(err)
			break
		}

		next, err := sp.processor(b[:n])
		if err != nil {
			goto CLOSE
		}

		if next != nil {
			sp.processor = next
			continue
		}

		// transport 开始双关通信
		remote, err := net.Dial("tcp", sp.remoteAddr)
		if err != nil {
			// 0x04 主机不可达
			sp.conn.Write([]byte{Socks5Version, 0x04, 0x00, AddressTypeIpv4})
			goto CLOSE
		}

		sp.conn.Write([]byte{Socks5Version, 0x00, 0x00, AddressTypeIpv4})
		sp.conn.Write(sp.remoteAddrRaw)

		log.Printf("连接目标服务 %s\n", sp.remoteAddr)

		st := NewSecureTransport(remote, sp.conn)
		st.Pip()
	}

CLOSE:
	log.Println("客户端连接段断开")
	sp.conn.Close()
}

// 握手协议一
// req [ver0x05, method_length, methods]
// res [0x05, 0x00]
func (sp *serverProtocol) handshake(b []byte) (processor, error) {
	if b[0] != Socks5Version {
		return nil, errors.New("非 socks5 握手协议一")
	}

	sp.conn.Write([]byte{Socks5Version, AuthNone})
	return sp.request, nil
}

// 握手协议二
// req [0x05, cmd0x01, 0x00, atype, dst.addr, dst.port]
// res [0x05, 0x00, 0x00, atype, bnd.addr, bnd.port]
func (sp *serverProtocol) request(b []byte) (processor, error) {
	cmd := b[1]

	if cmd != CmdConnect {
		return nil, errors.New("cmd 非 connect 请求")
	}

	atype := b[3]

	switch atype {
	case AddressTypeDomain:
		domainLength := b[4]

		domainRaw := b[5 : domainLength+5]
		domain := string(domainRaw)
		log.Printf("目标地址: %s\n", domain)
		addr, err := net.LookupHost(domain)
		if err != nil {
			return nil, errors.New("目标域名解析失败")
		}

		var port uint16
		portRaw := b[domainLength+5 : domainLength+7]
		binary.Read(bytes.NewBuffer(portRaw), binary.BigEndian, &port)
		sp.remoteAddr = fmt.Sprintf("%s:%d", addr[0], port)
		log.Printf("目标地址ip访问: %s\n", sp.remoteAddr)

		ipRaw := util.ConvertIP(addr[0])

		if ipRaw == nil {
			log.Println("目标地址ip异常")
		}
		sp.remoteAddrRaw = append(ipRaw, portRaw...)

	default:
		log.Println("当前请求目标地址格式暂未支持")
		return nil, errors.New("当前请求目标地址格式暂未支持")
	}

	return nil, nil
}

func NewServerProtocol(conn net.Conn) Protocol {
	sp := &serverProtocol{
		conn: conn,
	}

	sp.processor = sp.handshake

	return sp
}
