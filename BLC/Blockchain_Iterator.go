package BLC

import (
	"github.com/boltdb/bolt"
	"log"
)

//定义迭代器
type BlockchainIterator struct {
	CurrentHash  []byte      //当前正在遍历的区块的hash
	DB  *bolt.DB             //数据库对象
}

//迭代器遍历输出
func (blockchainIterator *BlockchainIterator)Next()*Block{
	var block *Block
	err := blockchainIterator.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlockBucket))
		if b != nil {
			currentBlockBytes := b.Get(blockchainIterator.CurrentHash)
			//获取到当前迭代器里面currentHash所对应的区块
			block = DeserializeBlock(currentBlockBytes)

			//更新迭代器里面的hash值
			blockchainIterator.CurrentHash = block.PrevBlockHash
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return block
}
