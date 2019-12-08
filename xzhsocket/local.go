package xzhsocket

//客户端，负责接收浏览器端给本地的数据，然后加密传送给服务器端。
//接收服务器端传回来的数据，解密返回给浏览器端
import (
	"net"
	"log"
)

type LsLocal struct {
	Cipher *cipher
	ListenAddr *net.TCPAddr
	RemoteAddr *net.TCPAddr 
}
//新建一个客户端
func NewLsLocal(password string, listenAddrS, remoteAddrS string) (*LsLocal, error) {
	bsPassword, err := ParsePassword(password)
	if err != nil {
		return nil, err
	}
	listenAddr, err := net.ResolveTCPAddr("tcp", listenAddrS)
	if err != nil {
		return nil, err 
	}
	remoteAddr, err := net.ResolveTCPAddr("tcp", remoteAddrS)
	if err != nil {
		return nil, err
	}
	return &LsLocal{
		Cipher: NewCipher(bsPassword),
		ListenAddr: listenAddr,
		RemoteAddr: remoteAddr, 
	}, nil 
}

func (local *LsLocal) Listen() error {
	return ListenSecureTCP(local.ListenAddr, local.Cipher, local.handleConn)
}

func (local *LsLocal) handleConn(userConn *SecureConn) {
	defer userConn.Close()
	log.Println("handling.....")
	//现在已经跟浏览器端也就是userconn建立了连接，现在要去连接远端了！
	//注意是加密的方式
	serverConn, err := DialSecureTCP(local.RemoteAddr, local.Cipher)
	if err != nil {
		log.Println(err.Error())
		return 
	}
	defer serverConn.Close()
	//连接已经建立好了，现在开始转发数据
	go func() {
		err = serverConn.DecodeCopy(userConn)//从服务端解密数据到浏览器端
		if err != nil {
			log.Println(err.Error())
			serverConn.Close()
			userConn.Close()
		}
	}()
	userConn.EncodeCopy(serverConn)// 浏览器端传数据给服务端，记得加密。

}