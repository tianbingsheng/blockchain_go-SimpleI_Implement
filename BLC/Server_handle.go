package BLC

import (
	"encoding/gob"
	"bytes"
	"log"
	"fmt"
)

//处理接收到的版本数据
func HandleVersion(request []byte, bc *Blockchain) {
	//1.从版本当中获取数据:[]byte
	//2.进行反序列化
	//3.操作bc获取自己的最后一个block的height
	//4.跟对方的进行比较
	versionBytes := request[COMMANDLENGTH:]

	//反序列化
	var version Version
	decoder := gob.NewDecoder(bytes.NewReader(versionBytes))
	err := decoder.Decode(&version)
	if err != nil {
		log.Panic(err)
	}

	//获取自己的block的bestHeight
	selfHeight := bc.Iterator().Next().Height //主节点自己的区块高度
	foreignerBestHeight := version.BestHeight //对方的区块高度

	fmt.Printf("接收到%s,传来的版本高度:%d\n", version.AddrFrom, foreignerBestHeight)

	//判断区块的高度情况  根据区块高度 进而同步区块链上的区块数据
	if selfHeight > foreignerBestHeight {
		//发送版本给对方
		SendVersion(version.AddrFrom, bc)
	} else if selfHeight < foreignerBestHeight {
		fmt.Println("我的高度没有你高,让我看看你的数据...........")
		SendGetBlocks(version.AddrFrom)
	}
}

//处理接收到的getblocks数据信息
func HandleGetBlocks(request []byte, blockchain *Blockchain) {

	getBlocksBytes := request[COMMANDLENGTH:]

	var getBlocks GetBlocks
	decoder := gob.NewDecoder(bytes.NewReader(getBlocksBytes))
	err := decoder.Decode(&getBlocks)
	if err != nil {
		log.Panic(err)
	}

	//打印数据的命令  以及来自谁的数据
	fmt.Printf("接收到了来自%s数据,%s\n", getBlocks.AddrFrom, string(request[:COMMANDLENGTH]))

	//查询自己的数据库,将区块的hash值进行拼接,发送给对方
	blocksHashes := blockchain.GetBlocksHashs()
	//发送自己的信息给对方

	SendInv(getBlocks.AddrFrom, BLOCK_TYPE, blocksHashes)
}

//处理接收到的Inv信息
func HandInv(request []byte, blockchain *Blockchain) {
	command := BytesToCommand(request[:COMMANDLENGTH])
	getInvBytes := request[COMMANDLENGTH:]

	var inv Inv
	decode := gob.NewDecoder(bytes.NewReader(getInvBytes))
	err := decode.Decode(&inv)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("接收到了来自%s的数据,%s\n", inv.AddrFrom, command)

	if inv.Type == BLOCK_TYPE {
		//获取对应数据的请求
		//当前的第一个区块  创世区块
		hash := inv.Items[0] //每个人的思路不同,可以进行区块的判断然后在进行添加
		SendGetData(inv.AddrFrom, BLOCK_TYPE, hash)

		//判断当前item的长度
		if len(inv.Items) >=1 {
			blocksArray=inv.Items[1:]
		}


	} else if inv.Type == TX_TYPE {

	}
}

//处理接收handleGetdata 的数据信息  对方向根据发来的hash值 索要对应的对象
func HandleGetData(request []byte, blockchain *Blockchain) {
	//处理对方发来的GetData
	command := BytesToCommand(request[:COMMANDLENGTH])
	getDataBytes := request[COMMANDLENGTH:]

	//反序列化
	var getData GetData
	decoder := gob.NewDecoder(bytes.NewReader(getDataBytes))
	err := decoder.Decode(&getData)
	if err != nil {
		log.Panic(err)
	}

	//打印接收到的数据
	fmt.Printf("接收到来自%s传来的命令,%s\n", getData.AddrFrom, command)

	//进行数据处理
	if getData.Type == BLOCK_TYPE {
		//根据getData中的hash值,来查找区块
		block := blockchain.GetBlockByHash(getData.Hash)
		SendBlock(getData.AddrFrom, block)

	} else if getData.Type == TX_TYPE {

	}
}

//处理获取的区块的数据
func HandBlockData(request []byte, blockchain *Blockchain) {
	command := BytesToCommand(request[:COMMANDLENGTH])
	getBlockDataBytes := request[COMMANDLENGTH:]

	var blockData BlockData
	decoder := gob.NewDecoder(bytes.NewReader(getBlockDataBytes))
	err := decoder.Decode(&blockData)

	if err != nil {
		log.Panic(err)
	}

	//打印信息数据的来源
	fmt.Printf("接收到来自%s的命令%s\n", blockData.AddrFrom, command)

	//取出block数据存入自己的数据库当中
	blockBytes := blockData.Block
	block := DeserializeBlock(blockBytes)

	//存入到本地的数据库
	blockchain.AddBlock(block)

	if len(blocksArray) == 0 {
		//重置UTXOSet表
		utxoSet := UTXOSet{blockchain}
		utxoSet.ResetUTXOSet()
	}
	if len(blocksArray) >0 {
		//发送请求,继续取获取
		SendGetData(blockData.AddrFrom,BLOCK_TYPE,blocksArray[0])
		blocksArray = blocksArray[1:]
	}
}




















