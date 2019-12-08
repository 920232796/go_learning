package xzhsocket

import (
	// "fmt"
	"net"
	"io"
	"log"
	"fmt"
)

const (
	bufSize = 1024
)

//安全的conn，里面包含了加密解密器
type SecureConn struct {
	Cipher *cipher
	io.ReadWriteCloser 
}

//先读一次，再把这一次数据解密
func (secureConn *SecureConn) DecodeRead(bs []byte) (n int, err error) {
	n, err = secureConn.Read(bs)
	if err != nil {
		return 
	}
	//开始解密
	secureConn.Cipher.decode(bs[:n])
	return n, nil 
}

//先加密一次数据，再写入到conn里面去
func (secureConn *SecureConn) EncodeWrite(bs []byte) (n int, err error) {
	secureConn.Cipher.encode(bs)
	n, err = secureConn.Write(bs)
	return n, err
}

//现在开始写 不断读数据的函数 从src里面不断读数据，解密写到dst里面
//比如客户端拿到服务端返回的加密数据之后，需要解密传给本地的浏览器端，第二种是服务端拿到客户端的加密数据之后，需要解密再进行请求
func (secureConn *SecureConn) DecodeCopy(dst io.ReadWriteCloser) error {
	buf := make([]byte, bufSize)
	for {
		readCount, readErr := secureConn.DecodeRead(buf)
		if readErr != nil {
			if readErr != io.EOF {
				return readErr 
			}
			return nil 
		}
		if readCount > 0 {
			writeCount, writeErr := dst.Write(buf[:readCount]) 
			if writeErr != nil {
				return writeErr
			}
			if readCount != writeCount {
				return io.ErrShortWrite
			}
		}
	}
}

//因为客户端要先加密浏览器给的数据，然后不断写给服务端,第二种是服务端得到数据，加密写回给客户端
//不断的读，然后加密，然后写到dst里面去
func (secureConn *SecureConn) EncodeCopy(dst io.ReadWriteCloser) error {
	buf := make([]byte, bufSize)
	for {
		readCount, readErr := secureConn.Read(buf)//这里直接读就ok，因为浏览器给的数据肯定不会加密
		if readErr != nil {
			if readErr != io.EOF {
				return readErr
			}
			return nil 
		}
		if readCount > 0 {
			//要去写到dst里面了，但是注意要进行加密再去写
			writeCount, writeErr := (&SecureConn{
				Cipher: secureConn.Cipher,
				ReadWriteCloser: dst,
			}).EncodeWrite(buf[:readCount])
			if writeErr != nil {
				return writeErr
			}
			if readCount != writeCount {
				return io.ErrShortWrite
			}
		}
	}
}

func ListenSecureTCP(laddr *net.TCPAddr, myCipher *cipher, handleConn func(localConn *SecureConn)) error {
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return err 
	}
	defer listener.Close()
	//开始监听
	fmt.Println("listening....")
	for {
		localConn, err := listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue 
		}
		localConn.SetLinger(0)
		//对于不通的端，有不同的处理函数哈
		go handleConn(&SecureConn{
			Cipher: myCipher, 
			ReadWriteCloser: localConn,
		})
	}
}

func DialSecureTCP(raddr *net.TCPAddr, cipher *cipher) (*SecureConn, error) {
	serverConn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return nil, err
	}
	return &SecureConn{
		Cipher: cipher,
		ReadWriteCloser: serverConn,
	}, nil 

}