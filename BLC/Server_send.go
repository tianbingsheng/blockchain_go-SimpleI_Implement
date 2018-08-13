package BLC

import (
	"fmt"
	"net"
	"log"
	"io"
	"bytes"
)

//发送消息
//发送的数据格式的判定   版本数据 区块数据 交易数据  等等  需要进行命令的处理
/*
命令:
	data: command + 结构体数据
          version + 版本数据
          blocks  + 区块的数据

          data  []byte
*/

//处理节点之间发送的数据

//发送version信息
func SendVersion(toAddr string, blockchain *Blockchain) {
	//获取当前区块链的最高高度
	bestHeight := blockchain.GetBestHeight()
	//版本的结构体信息,版本对象
	version := Version{NODE_VERSION, bestHeight, nodeAddress}
	//将版本对象序列化
	payload := gobEncode(version)
	//拼接命令加上数据
	request := append(commandToBytes(COMMAND_VERSION), payload...)

	//发送数据
	SendData(toAddr, request)
}

//发送数据的格式  command + version结构体数据
/*
规定COMMANDLENGTH = 12  前12字节的数据是存放命令的格式  来制定你要发送的数据是什么?
*/
//获取区块信息
func SendGetBlocks(toAddr string) {

	getBlocks := GetBlocks{nodeAddress}
	payload := gobEncode(getBlocks)
	//数据拼接
	request := append(commandToBytes(COMMAND_GETBLOCKS), payload...)
	SendData(toAddr, request)
}

func SendData(to string, data []byte) {
	fmt.Println("当前节点可以发送数据....")
	conn, err := net.Dial("tcp", to)

	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	//发送数据
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}

}

//
func SendInv(toAddr string, kind string, data [][]byte) {
	inv := Inv{nodeAddress, kind, data}
	//拼接要发送的数据

	payload := gobEncode(inv)
	request := append(commandToBytes(COMMAND_INV), payload...)
	SendData(toAddr, request)
}

//发送要请求数据的命令
func SendGetData(toAddr string, kind string, hash []byte) {
	getData := GetData{nodeAddress, kind, hash}
	payload := gobEncode(getData)
	request := append(commandToBytes(COMMAND_GETDATA), payload...)
	SendData(toAddr, request)
}

//发送区块
func SendBlock(toAddr string, block *Block) {

	blockData := BlockData{nodeAddress, block.Serialize()}
	payload := gobEncode(blockData)
	request := append(commandToBytes(COMMAND_BLOCKDATA), payload...)
	SendData(toAddr, request)
}
