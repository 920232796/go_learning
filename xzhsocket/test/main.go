package main 

import (
	"fmt"
	"xzhsocket"
	"os"
	// "cipher/test"
)

//一些测试函数
func main() {
	fmt.Println("hello world")
	///测试生成密码
	bytePassword, err := xzhsocket.ParsePassword(xzhsocket.RandPassword())
	if err != nil {
		fmt.Println(err.Error())
		return 
	}
	for _, v := range bytePassword {
		fmt.Print(v)
		fmt.Print(" ")
	}
	fmt.Println()

	//生成一个密码就ok
	myPassword := xzhsocket.RandPassword()
	fmt.Println("length: ", len(myPassword))
	pwf, err := os.OpenFile("./password.txt", os.O_CREATE|os.O_RDWR, 0666)
	defer pwf.Close()
	if err != nil {
		fmt.Println(err.Error())
		// fmt.Println("dsaddsa")
		return 
	}
	n, err := pwf.WriteString(myPassword)
	// fmt.Println("123123123132", myPassword)
	// n, err := pwf.Write(bytePassword[:])
	if err != nil {
		fmt.Println(err.Error())
		return 
	}
	fmt.Println(n)

	//测试生成 编码解码器
	bsMyPassword, _ := xzhsocket.ParsePassword(myPassword)
	myCipher := xzhsocket.NewCipher(bsMyPassword)
	for i, v := range myCipher.EncodePassword {
		if byte(i) == myCipher.DecodePassword[v] {
			continue
		} else {
			fmt.Println("err!")
			return
		}
	}
	//运行到这说明加密解密器cipher没问题
	fmt.Println("cipher is right ")

	//读一下密码。看看是不是读出来是原来写入的字符串
	xzhsocket.ReadPassword()

}