package BLC

import (
	"fmt"
	"net"
	"log"
	"io/ioutil"
)

/*
全节点:地址硬编码
		localhost:3000
钱包节点 :
		localhost:3001
矿工节点:
		localhost:3002
*/

//节点服务启动,可以去接受其他节点的传来的数据
func startServer(nodeID string, minerAddress string) {
	//当前节点自己的地址
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)

	//监听:地址
	listen, err := net.Listen("tcp", nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer listen.Close()

	//当前节点的判断,如果是全节点:等待
	//如果不是全节点  ,钱包节点  或者是矿工节点,给全节点发送一个消息....
	blockchain := BlockchainObject(nodeID)

	//不是主节点,会同步主节点的数据信息
	if nodeAddress != KonwNodes[0] {
		//钱包节点,矿工节点.........给全节点发送一个消息..
		//SendMessage(KonwNodes[0],"我是王二狗:"+nodeAddress)
		SendVersion(KonwNodes[0], blockchain)

	}

	for {
		conn, err := listen.Accept() //会发生阻塞
		if err != nil {
			log.Panic(err)
		}

		fmt.Println("发送方已经连入", conn.RemoteAddr())

		go HandleConnection(conn,blockchain)
	}
}

//处理不同协程之间客户端信息处理
func HandleConnection(conn net.Conn,bc *Blockchain) {
	//读取对方传来的数据
	request, err := ioutil.ReadAll(conn) // command + 数据
	if err != nil {
		log.Panic(err)
	}

	command := BytesToCommand(request[:COMMANDLENGTH])

	fmt.Printf("接收到的命令是:%s\n", command)

	fmt.Println(command)
	//根据命令,做出相对应响应的命令
	switch command {
	case COMMAND_VERSION:
		//此处是处理接收到的版本的数据
		HandleVersion(request,bc)
	case COMMAND_GETBLOCKS:
		//此处处理接收等到的blocks数据
		HandleGetBlocks(request,bc)
	case COMMAND_INV:
		//此处处理的是收到的Inv信息
		HandInv(request,bc)
	case COMMAND_GETDATA:
		//对方发来Getdata的信息
		HandleGetData(request,bc)
	case COMMAND_BLOCKDATA:
		//接收到blockdata的数据:发来真正的区块
		HandBlockData(request,bc)
	default:
		fmt.Println("----------您输入的命令不符合要求---------")
	}

	defer conn.Close()
}















