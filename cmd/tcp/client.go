package main

import (
	"io"
	"log"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", ":2020")
	if err != nil {
		panic(err)
	}
	done := make(chan struct{})
	go func() {
		io.Copy(os.Stdout, conn) // 注意：忽略错误
		log.Println("done")
		done <- struct{}{} // 服务器的done
	}()
	mustCopy(conn, os.Stdin) // 标准输入 -> 服务器链接
	conn.Close()
	<-done // 等待服务器的goroutine
}

func mustCopy(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
	}
}

// 新开的goroutine通过一个channel和main goroutine通信
