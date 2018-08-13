package BLC

import (
	"fmt"
	"os"
)

func (cli *CLI)StartNode(nodeID string,minerAddress string){
	//启动服务器  go语言这是一个逻辑或 符合条件后面就不判断了
	if minerAddress == "" || IsValidForAddress([]byte(minerAddress)) {
		//启动服务器
		startServer(nodeID,minerAddress)
	}else{
		//地址无效
		fmt.Println("地址无效")
		os.Exit(1)
	}
}