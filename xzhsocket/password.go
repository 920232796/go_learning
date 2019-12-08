package xzhsocket

//用来生成密码

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
	"errors"
	"encoding/base64"
)

const passwordLength int = 256

type password [passwordLength]byte

func init() {
	//随机种子
	rand.Seed(time.Now().Unix())
}

//password类型转换为字符串
func (password *password) String() string {
	return base64.StdEncoding.EncodeToString(password[:]) //把密码 字节类型，转换为字符串类型
}

//把密码字符串解析为password类型～
func ParsePassword(passwordString string) (*password, error) {
	bs, err := base64.StdEncoding.DecodeString(strings.TrimSpace(passwordString)) //去除一下多余空格
	if err != nil || len(bs) != passwordLength {
		fmt.Println("parse password err is ", err.Error())
		return nil, errors.New("不合法的密码")
	}
	password := password{}

	copy(password[:], bs)
	bs = nil 
	return &password, nil 
}

//生成密码
func RandPassword() string {

	array := rand.Perm(passwordLength)
	password := password{}

	for i, v := range array {
		password[i] = byte(v)
		if i == v {
			return RandPassword()
		}
	}
	fmt.Println(password)
	return password.String()
}
