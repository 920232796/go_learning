package xzhsocket

import (
	"encoding/binary"
	"net"
	"fmt"
)

//新建一个服务端，监听客户端发来的消息，解码，解析socket协议
type LsServer struct {
	Cipher *cipher
	ListenAddr *net.TCPAddr

}

func NewLsServer(password string, listenAddr string) (*LsServer, error){
	bsPassword, err := ParsePassword(password)
	if err != nil {
		// fmt.Println(err.Error())
		return nil, err
	}

	TCPAddr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return nil, err
	}

	return &LsServer{
		Cipher: NewCipher(bsPassword),
		ListenAddr: TCPAddr,
	}, nil 

}

//服务端启动监听
func (lsServer *LsServer) Listen() error {
	return ListenSecureTCP(lsServer.ListenAddr, lsServer.Cipher, lsServer.handleConn)
}

//服务端的处理函数 每监听到一次客户端的请求，都会调用handle处理函数。
func (lsServer *LsServer) handleConn(localConn *SecureConn) {
	defer localConn.Close()

	buf := make([]byte, 256)
	/**
	   The localConn connects to the dstServer, and sends a ver
	   identifier/method selection message:
		          +----+----------+----------+
		          |VER | NMETHODS | METHODS  |
		          +----+----------+----------+
		          | 1  |    1     | 1 to 255 |
		          +----+----------+----------+
	   The VER field is set to X'05' for this ver of the protocol.  The
	   NMETHODS field contains the number of method identifier octets that
	   appear in the METHODS field.
	*/
	// 第一个字段VER代表Socks的版本，Socks5默认为0x05，其固定长度为1个字节
	_, err := localConn.DecodeRead(buf)//读出来并且解密数据 放到buf里面，这个数据是先握手相当于，客户端问服务端验证方式是啥
	// 只支持版本5
	if err != nil || buf[0] != 0x05 {
		return
	}
	fmt.Println(buf[0])

	/**
	   The dstServer selects from one of the methods given in METHODS, and
	   sends a METHOD selection message:

		          +----+--------+
		          |VER | METHOD |
		          +----+--------+
		          | 1  |   1    |
		          +----+--------+
	*/
	// 不需要验证，直接验证通过
	localConn.EncodeWrite([]byte{0x05, 0x00})//把这两个byte的数据写入到流中，也就是告诉客户端，不需要进行验证

	/**
	  +----+-----+-------+------+----------+----------+
	  |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	  +----+-----+-------+------+----------+----------+
	  | 1  |  1  | X'00' |  1   | Variable |    2     |
	  +----+-----+-------+------+----------+----------+
	*/
	// 获取真正的远程服务的地址
	n, err := localConn.DecodeRead(buf)// buf大小为256个字节，因此读256个字节
	// n 最短的长度为7 情况为 ATYP=3 DST.ADDR占用1字节 值为0x0
	if err != nil || n < 7 {
		return
	}

	// CMD代表客户端请求的类型，值长度也是1个字节，有三种类型
	// CONNECT X'01'
	if buf[1] != 0x01 {
		// 目前只支持 CONNECT
		return
	}

	var dIP []byte
	// aType 代表请求的远程服务器地址类型，值长度1个字节，有三种类型
	switch buf[3] {
	case 0x01:
		//	IP V4 address: X'01'
		dIP = buf[4 : 4+net.IPv4len]
	case 0x03:
		//	DOMAINNAME: X'03'
		ipAddr, err := net.ResolveIPAddr("ip", string(buf[5:n-2]))
		if err != nil {
			return
		}
		dIP = ipAddr.IP
	case 0x04:
		//	IP V6 address: X'04'
		dIP = buf[4 : 4+net.IPv6len]
	default:
		return
	}
	dPort := buf[n-2:]
	dstAddr := &net.TCPAddr{
		IP:   dIP,
		Port: int(binary.BigEndian.Uint16(dPort)),
	}
	fmt.Println("dst addr is :", dstAddr)
	// fmt.Println(dstAddr)
	
	// 连接真正的远程服务，这时候不需要加密啥的了 直接连接就ok
	dstServer, err := net.DialTCP("tcp", nil, dstAddr)
	if err != nil {
		return
	} else {
		defer dstServer.Close()
		// Conn被关闭时直接清除所有数据 不管没有发送的数据
		dstServer.SetLinger(0)
		// 响应客户端连接成功
		/**
		  +----+-----+-------+------+----------+----------+
		  |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
		  +----+-----+-------+------+----------+----------+
		  | 1  |  1  | X'00' |  1   | Variable |    2     |
		  +----+-----+-------+------+----------+----------+
		*/
		// 响应客户端连接成功，这是服务端告诉客户端这些信息，说明连接成功了！
		localConn.EncodeWrite([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}

	//开始传输数据了,从客户端读取数据，解密，发送到目标服务器
	go func() {
		err := localConn.DecodeCopy(dstServer) //本地端源源不断的读数据，解密到dst
		if err != nil {
			localConn.Close()
			dstServer.Close()
		}
	}()
	//从目标服务器得到返回数据，加密，传输到本地端！
	(&SecureConn{
		Cipher: localConn.Cipher,
		ReadWriteCloser: dstServer,
	}).EncodeCopy(localConn) //得到dst返回的数据，加密写回到本地端!

}

