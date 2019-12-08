package xzhsocket

//读一下密码文件！

import (
	"fmt"
	"os"
	"io"

)

func ReadPassword() {

	fmt.Println("hello world")

	pwd, err := os.OpenFile("./password.txt", os.O_RDWR, 0666)
	defer pwd.Close()
	if err != nil {
		fmt.Println(err.Error())
		return 
	}

	buf := make([]byte, 512)
	readCount, err := pwd.Read(buf)
	if err != nil {
		if err == io.EOF {
			return 
		}
		fmt.Println(err.Error())
		return 
	}

	fmt.Println(readCount)
	fmt.Println(string(buf[:readCount]))

	pass, err := ParsePassword(string(buf[:readCount]))
	if err != nil {
		fmt.Println(err.Error())
		return 
	}
	fmt.Println(pass[:])
}