package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

// 使用通信来共享内存
var ( // 新用户到来，通过该channel进行登记 如果不是有缓存的channel，这里会不会阻塞呢？
	// 回答：会阻塞，因为没有人来接收这个消息
	// 那这里为什么还用无缓存的channel呢？
	// 因为这里只是登记，不需要缓存
	enteringChannel = make(chan *User)
	// 用户离开，通过该channel进行登记
	leavingChannel = make(chan *User)
	// 广播消息
	messageChannel = make(chan string, 8)
)

func main() {
	lister, err := net.Listen("tcp", ":2020")
	if err != nil {
		panic(err)
	}
	go broadcaster()
	for {
		conn, err := lister.Accept()
		if err != nil {
			continue
		}
		go handleConn(conn) // 相当于是来一个请求就开一个协程
		// 但是这样会导致资源的浪费
		// 还有个问题就是,如果有一个用户发送消息,那么所有的用户都会收到消息
		// 但是我们的需求是,只有自己发送的消息才能收到
		// 但是如果是协程池来处理的话，会不会导致阻塞呢？
	}
}

// 广播给所有的用户 用户记录聊天室用户，并进行消息广播
func broadcaster() {
	users := make(map[*User]struct{})
	for {
		select {
		case user := <-enteringChannel:
			// 新用户进入
			users[user] = struct{}{}
		case user := <-leavingChannel:
			// 用户离开
			delete(users, user)
			// 关闭用户的通道
			close(user.MessageChannel)
		case msg := <-messageChannel:
			// 给所有的在线用户发送消息
			for user := range users {
				user.MessageChannel <- msg
			}
		}
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	// 每一个连接处理完要进行关闭
	// 新用户进来,构建用户的实例
	user := &User{
		ID:             GetUserID(),
		Addr:           conn.RemoteAddr().String(),
		EnterAt:        time.Now(),
		MessageChannel: make(chan string, 8),
	}
	// 由于当前是在一个新的协程中进行读操作的，所以需要开一个协程进行写操作，读写goroutine通过channel进行通信
	go SendMessage(conn, user.MessageChannel)
	// 向该用户发送欢迎消息，并通知其他用户新用户的到来
	user.MessageChannel <- "Welcome, " + user.Addr
	messageChannel <- user.Addr + " has arrived"
	// 将新用户登记到enteringChannel中 记录到全局用户里面,避免用锁   channel + map 可以避免用锁
	enteringChannel <- user
	// 循环读取用户输入
	input := bufio.NewScanner(conn)
	for input.Scan() {
		messageChannel <- user.Addr + ": " + input.Text()
	}
}

// User 用户实例
type User struct {
	ID             int         // 用户的唯一标识
	Addr           string      // 用户的IP和端口
	EnterAt        time.Time   // 用户进入的时间
	MessageChannel chan string // 用户发送的消息
}

// SendMessage 用户发送消息
func SendMessage(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}

}

func GetUserID() int {
	return 1
}

// 协程之间通过channel去通信
// 远程医疗直播平台

// 用map去存储在线用户

// 1.理解什么是网络流
// 就是一堆字节流，这些字节流是通过网络传输的，所以叫网络流
// 就是一些数据 通过网络传输的

// 设置日志信息的唯一标识 打印成json格式 并收集起来 用于分析
// cmd 取名的含义
// 如果是我后台设置， 那就是那就是通过error信息去找唯一ID
