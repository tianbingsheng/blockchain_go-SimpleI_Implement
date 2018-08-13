package main

import (
	"publicChain/part76-Net-Handle_Message2/BLC"
)

func main() {
	cli := &BLC.CLI{}
	cli.Run()
}

//达到的效果是执行./main 会创建创世区块
// 然后根据所输入操作执行相应的步骤