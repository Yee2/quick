// +build !windows

package main

import (
	"net"
	"syscall"
	"fmt"
	"errors"
	"io"
	"encoding/binary"
	"strconv"
	"github.com/lucas-clemente/quic-go"
)

const (
	SO_ORIGINAL_DST      = 80
	IP6T_SO_ORIGINAL_DST = 80
)

func direct(conn net.Conn,session quic.Session) error {
	defer conn.Close()
	TCPConn,ok := conn.(*net.TCPConn)
	if !ok{
		return errors.New("连接不是TCP连接")
	}
	address,local,err := getOriginalDstAddr(TCPConn)
	if err != nil{
		return err
	}
	defer local.Close()
	server,err := s5dialer(session,address)
	if err != nil{
		return err
	}
	defer server.Close()
	go func() {
		_, e := io.Copy(local, server)
		if e != nil {
			err = e
		}
	}()
	_, e := io.Copy(server, local)
	if e != nil {
		err = e
	}
	return err
}

func s5dialer(session quic.Session,target *net.TCPAddr) (stream quic.Stream,err error ){
	defer func() {
		if e := recover(); e != nil{
			err = fmt.Errorf("socks5 proxy error:%s",e)
			if stream != nil{
				stream.Close()
				stream = nil
			}
		}
	}()
	stream,err = session.OpenStream()
	die(err)

	stream.Write([]byte{0x05,0x01,0x00})
	var raw [1024]byte
	_,err = stream.Read(raw[:])
	die(err)
	if binary.BigEndian.Uint16(raw[:2]) != 0x0500{
		panic(errors.New("doesn't supported this proxy server"))
	}
	p := make([]byte,2,2)
	binary.BigEndian.PutUint16(p, uint16(target.Port))
	_,err = stream.Write(append(append([]byte{0x05,0x01,0x00,0x01},[]byte(target.IP.To4())...),p...))
	die(err)
	_,err = stream.Read(raw[:])
	if err == io.EOF {
		panic(errors.New("connection closed by remote server"))
	}
	die(err)
	if raw[0] != 0x05 || raw[2] != 0x00{
		panic(errors.New("不支持代理服务器！"))
	}

	switch raw[1] {
	case 0x00:
		break
	case 0x01:
		panic(errors.New("服务器错误:X'01' general SOCKS server failure"))
	case 0x02:
		panic(errors.New("服务器错误:X'02' connection not allowed by ruleset"))
	case 0x03:
		panic(errors.New("服务器错误:X'03' Network unreachable"))
	case 0x04:
		panic(errors.New("服务器错误:X'04' Host unreachable"))
	case 0x05:
		panic(errors.New("服务器错误:X'05' Connection refused"))
	case 0x06:
		panic(errors.New("服务器错误:X'06' TTL expired"))
	case 0x07:
		panic(errors.New("服务器错误:X'07' Command not supported"))
	case 0x08:
		panic(errors.New("服务器错误:X'08' Address type not supported"))
	case 0x09:
		panic(errors.New("服务器错误:X'09' to X'FF' unassigned"))
	default:
		panic(errors.New("unknown code: " + strconv.Itoa(int(raw[1]))))
	}
	switch raw[3] {
	case 0x01:
		// TODO: 连接构建完成
		return stream,nil
	case 0x03:
		// TODO： 绑定的是域名
	case 0x04:
		// TODO： 绑定的是IPv6地址
	default:
		// TODO： 未知类型
		panic(errors.New("未知代理类型"))
	}

	return nil,errors.New("不支持代理类型")
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
	ip := net.IPv4(mreq.Multiaddr[4], mreq.Multiaddr[5], mreq.Multiaddr[6], mreq.Multiaddr[7])
	port := uint16(mreq.Multiaddr[2])<<8 + uint16(mreq.Multiaddr[3])
	addr, err = net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", ip.String(), port))
	if err != nil {
		return
	}

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
