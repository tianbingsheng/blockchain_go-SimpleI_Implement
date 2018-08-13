package BLC

import (
	"time"
	"fmt"
	"bytes"
	"encoding/gob"
	"log"
)

type Block struct {
	Height        int64          //1.区块高度
	PrevBlockHash []byte         //2.上一个区块的hash值
	Txs           []*Transaction //3.交易数据		//一个区块里面可以打包多笔交易
	Timestamp     int64          //4.时间戳
	Hash          []byte         //5.Hash
	Nouce         int64          //随机值
}
//需要将Txs转换成字节数组[]byte
func (block *Block)HashTransactions()[]byte{
	//只需要把每笔交易的hash值拼接起来就成,不需要把所有字段进行拼接
	//利用merkle树,可以追溯交易列表来源

	//var txHashes [][]byte
	//var txHash [32]byte
	//for _,tx := range block.Txs{
	//	txHashes = append(txHashes,tx.TxHash)
	//}
	////字节数组的拼接,可以有不同的处理方式
	//txHash = sha256.Sum256(bytes.Join(txHashes,[]byte{}))
	//return txHash[:]
	var txs [][]byte
	for _,tx := range block.Txs{
		txs = append(txs,tx.Serialize())
	}
	mTree := NewMerkleTree(txs)
	return mTree.RootNode.DataHash
}

//将区块对象序列化为字节数组
func (block *Block) Serialize() []byte {
	//创建缓冲区
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

//反序列化
func DeserializeBlock(blockBytes []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	err := decoder.Decode(&block)
	if err != nil {
		log.Fatal(err)
	}
	return &block
}

//1.创建新的区块
func NewBlock(txs []*Transaction, height int64, prevBlockHash []byte) *Block {
	//创建区块
	block := &Block{
		Height:        height,
		PrevBlockHash: prevBlockHash,
		Txs:           txs,
		Timestamp:     time.Now().Unix(),
		Hash:          nil,
		Nouce:         0,
	}
	//调用工作量证明方法并且返回有效的hash和Nouce值
	pow := NewProofOfWork(block)
	//挖矿验证
	hash, nouce := pow.Run()

	block.Hash = hash[:]
	block.Nouce = nouce
	fmt.Println()
	return block
}

//2.单独写一个方法,生成创世区块
func CreateGenesisBlock(txs []*Transaction) *Block {
	return NewBlock(txs, 0, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}
