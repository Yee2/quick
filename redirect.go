// +build !windows

package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"syscall"
)

const (
	SO_ORIGINAL_DST      = 80
	IP6T_SO_ORIGINAL_DST = 80
)

func direct(conn net.Conn, stream io.ReadWriteCloser) error {
	defer conn.Close()
	TCPConn, ok := conn.(*net.TCPConn)
	if !ok {
		return errors.New("connection is not a TCP connection")
	}
	address, local, err := getOriginalDstAddr(TCPConn)
	if err != nil {
		return err
	}
	defer local.Close()
	server, err := s5dialer(stream, address)
	if err != nil {
		return err
	}
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		io.Copy(local, server)
		cancel()
	}()
	go func() {
		io.Copy(server, local)
		cancel()
	}()
	<-ctx.Done()
	return err
}

func s5dialer(stream io.ReadWriteCloser, target *net.TCPAddr) (Socks5Bind io.ReadWriteCloser, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("socks5 proxy error:%s", e)
		}
	}()
	var raw [1024]byte
	// handshake : client hello
	stream.Write([]byte{0x05, 0x01, 0x00})
	_, err = stream.Read(raw[:])
	die(err)
	if binary.BigEndian.Uint16(raw[:2]) != 0x0500 {
		panic(errors.New("proxy server requires authentication and does not support authentication"))
	}

	// socks5 stage 2
	p := make([]byte, 2, 2)
	binary.BigEndian.PutUint16(p, uint16(target.Port))
	if ip := target.IP.To4(); ip != nil {
		// 目标地址是 IPv4
		_, err = stream.Write([]byte{0x05, 0x01, 0x00, 0x01, ip[0], ip[1], ip[2], ip[3], p[0], p[1]})
		die(err)
	} else if ip := target.IP.To16(); ip != nil {
		// 目标地址是 IPv6
		_, err = stream.Write(append(append([]byte{0x05, 0x01, 0x00, 0x04}, ip...), p...))
		die(err)
	} else {
		panic(errors.New("wrong destination address"))
	}

	_, err = stream.Read(raw[:])
	if err == io.EOF {
		panic(errors.New("connection closed by remote server"))
	}
	die(err)
	if raw[0] != 0x05 || raw[2] != 0x00 {
		panic(errors.New("不支持代理服务器！"))
	}

	switch raw[1] {
	case 0x00:
		break
	case 0x01:
		panic(errors.New("connection failed:0x01 general SOCKS server failure"))
	case 0x02:
		panic(errors.New("connection failed:0x02 connection not allowed by ruleset"))
	case 0x03:
		panic(errors.New("connection failed:0x03 Network unreachable"))
	case 0x04:
		panic(errors.New("connection failed:0x04 Host unreachable"))
	case 0x05:
		panic(errors.New("connection failed:0x05 Connection refused"))
	case 0x06:
		panic(errors.New("connection failed:0x06 TTL expired"))
	case 0x07:
		panic(errors.New("connection failed:0x07 Command not supported"))
	case 0x08:
		panic(errors.New("connection failed:0x08 Address type not supported"))
	default:
		panic(errors.New("connection failed:unknown code: " + strconv.Itoa(int(raw[1]))))
	}
	switch raw[3] {
	case 0x01:
		// TODO: 连接构建完成
		return stream, nil
	case 0x03:
		// TODO： 绑定的是域名
		panic(errors.New("unexpected error"))
	case 0x04:
		// TODO： 绑定的是IPv6地址
		panic(errors.New("unexpected error"))
	default:
		// TODO： 未知类型
		panic(errors.New("unexpected error"))
	}
}

func getOriginalDstAddr(conn *net.TCPConn) (addr *net.TCPAddr, c *net.TCPConn, err error) {
	defer conn.Close()

	fc, err := conn.File()
	if err != nil {
		return
	}
	defer fc.Close()

	mreq, err := syscall.GetsockoptIPv6Mreq(int(fc.Fd()), syscall.IPPROTO_IP, IP6T_SO_ORIGINAL_DST)
	if err != nil {
		return
	}

	// only ipv4 support
	addr = &net.TCPAddr{IP: mreq.Multiaddr[4:8], Port: int(mreq.Multiaddr[2])<<8 + int(mreq.Multiaddr[3])}

	cc, err := net.FileConn(fc)
	if err != nil {
		return
	}

	c, ok := cc.(*net.TCPConn)
	if !ok {
		err = errors.New("not a TCP connection")
	}
	return
}
