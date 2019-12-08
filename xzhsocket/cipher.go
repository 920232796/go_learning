package xzhsocket
//负责加密的一些包
import (
	// "io"
)

type cipher struct {
	EncodePassword *password 
	DecodePassword *password 
}

//加密byte流
func (cipher *cipher) encode(bs []byte) {
	for i, v := range bs {
		bs[i] = cipher.EncodePassword[v]
	}
}
//解密byte流
func (cipher *cipher) decode(bs []byte) {
	for i, v := range bs {
		bs[i] = cipher.DecodePassword[v]
	}
}

//新建一个编码解码器
func NewCipher(encodePassword *password) *cipher {
	decodePassword := &password{}
	for i, v := range encodePassword {
		encodePassword[i] = v
		decodePassword[v] = byte(i) //反过来就ok了 解码器
	}
	return &cipher{
		EncodePassword: encodePassword,
		DecodePassword: decodePassword,
	}
}