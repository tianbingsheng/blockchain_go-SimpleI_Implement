package BLC

import (
	"github.com/boltdb/bolt"
	"log"
	"fmt"
	"math/big"
	"time"
	"os"
	"strconv"
	"encoding/hex"
	"crypto/ecdsa"
	"bytes"
)

type Blockchain struct {
	Tip []byte   //最新的区块的hash值(始终保持的是最新区块的hash)
	DB  *bolt.DB //数据库对象
}

//迭代器
func (blockchain *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{blockchain.Tip, blockchain.DB}
}

//1.创建带有创世区块的区块链
func CreateBlockchainWithGenesis(address string, node_ID string) *Blockchain {
	//判断数据库是否存在
	//设置拼接新的数据库名字
	DBName := fmt.Sprintf(dbName, node_ID)

	if DBExists(DBName) {
		fmt.Println("创世区块已经存在...")
		os.Exit(1)
	}

	fmt.Println("--------正在创建创世区块--------")
	//创建或者打开数据库

	db, err := bolt.Open(DBName, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	//关闭数据库
	//defer db.Close()

	//创建创世区块
	//创建了一个coinBase交易
	txCoinBase := NewCoinBaseTransaction(address)

	//需要传入的是交易列表的数组
	genesisBlock := CreateGenesisBlock([]*Transaction{txCoinBase})

	err = db.Update(func(tx *bolt.Tx) error {

		//创建桶
		b, err := tx.CreateBucket([]byte(BlockBucket))
		if err != nil {
			log.Panic(err)
		}

		//如果桶存在
		if b != nil {
			//将创世区块存储到桶当中
			err = b.Put(genesisBlock.Hash, genesisBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}
			//存储最新的区块的hash
			err = b.Put([]byte("l"), genesisBlock.Hash)
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return &Blockchain{genesisBlock.Hash, db}
}

//增加区块到区块链当中(增加的新区块放在数据库持久化中)
func (blc *Blockchain) AddBlockToBlockchain(txs []*Transaction) *Blockchain {

	err := blc.DB.Update(func(tx *bolt.Tx) error {
		//1.获取桶
		b := tx.Bucket([]byte(BlockBucket))
		//获取最新的区块
		blockBytes := b.Get(blc.Tip)
		oldBlock := DeserializeBlock(blockBytes)

		//2.创建新区快
		newBlock := NewBlock(txs, oldBlock.Height+1, oldBlock.Hash)

		//3将区块序列化并且存储到数据库当中
		b.Put(newBlock.Hash, newBlock.Serialize())
		//4.更新数据库里面的"l"对应的hash
		b.Put([]byte("l"), newBlock.Hash)
		//5.更新blockchain的Tip(存的是最新的区块hash值)
		blc.Tip = newBlock.Hash
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	return blc
}

//判断当前数据库是否存在
func DBExists(dbName string) bool {
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		return false
	}
	return true
}

//遍历输出所有的区块的信息
func (blc *Blockchain) Printchain() {
	blockchainIterator := blc.Iterator()
	var block *Block
	for {
		block = blockchainIterator.Next()

		fmt.Printf("Height：%d\n", block.Height)
		fmt.Printf("PrevBlockHash：%x\n", block.PrevBlockHash)
		//将时间戳固定格式输出的形式
		fmt.Printf("Timestamp：%s\n", time.Unix(block.Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash：%x\n", block.Hash)
		fmt.Printf("Nonce：%d\n", block.Nouce)
		fmt.Println("Txs:")
		//遍历交易数据
		for _, tx := range block.Txs {
			fmt.Printf("%x\n", tx.TxHash)

			fmt.Println("Vins")
			for _, in := range tx.Vins {
				fmt.Printf("%x\n", in.TXHash)
				fmt.Printf("%d\n", in.Vout)
				fmt.Println(in.PublicKey)
			}

			fmt.Println("Vouts")
			for _, out := range tx.Vouts {
				fmt.Println(out.Value)
				fmt.Println(out.Ripemd160Hash)
			}
		}

		fmt.Println("-------------------------------------------------------------")

		var hashInt *big.Int
		hashInt = big.NewInt(1)
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(hashInt) == 0 {
			break
		}
	}
}

//获取blockchain Object
func BlockchainObject(nodeid string) *Blockchain {
	//打开数据库
	DBName := fmt.Sprintf(dbName, nodeid)
	db, err := bolt.Open(DBName, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	var tip []byte
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlockBucket))
		if b != nil {
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return &Blockchain{tip, db}
}

//如果一个地址所对应的txOutput未花费,则这个Transaction就应该添加到数组中返回
func (blockchain *Blockchain) UnUTXOs(address string, txs []*Transaction) []*UTXO {

	// var unUTXOs []*TXOutput

	var unUTXOs []*UTXO //要保存记录他的交易hash 索引
	//记录已经花费的输出
	spentTXOutput := make(map[string][]int)
	//{hash :[0,1...]}

	//首先查询本次转账已经创建的transaction 然后在查询数据库
	//--------首先查询还没有上区块的交易-----------------
	for i := len(txs) - 1; i >= 0; i-- {
		tx := txs[i]
		//txHash

		//Vins
		//保证他不是coinbase交易   //主要提高效率,直接忽略遍历coinbase交易
		if !tx.IsCoinbaseTransaction() {
			for _, in := range tx.Vins {
				//查看已经花费的,是否已经解锁
				//获取公钥hash
				version_pubKey_checksum := Base58Decode([]byte(address))
				pubKeyHash := version_pubKey_checksum[1:len(version_pubKey_checksum)-4]
				if in.UnlockRipemd160(pubKeyHash) {
					//十六进制字符串
					key := hex.EncodeToString(in.TXHash)
					//主要是相同的交易hash可能存在多条输出
					spentTXOutput[key] = append(spentTXOutput[key], in.Vout)
				}
			}
		}

		//Vouts
		//遍历出来  包含已经花费和未花费
		//定义标签名字:
	outputs1:
		for index, out := range tx.Vouts {
			if out.UnlockScriptPubKeyWithAddress(address) {
				//如果已花费的map不为空
				if len(spentTXOutput) != 0 {
					var isTxOutput bool //false   首先默认是未花费

					for txHash, indexArray := range spentTXOutput {

						for _, i := range indexArray {
							if index == i && txHash == hex.EncodeToString(tx.TxHash) {
								isTxOutput = true
								continue outputs1
							}
						}
					}
					//将未花费的记录output添加到数组当中
					if !isTxOutput {
						utxo := &UTXO{tx.TxHash, index, out}
						unUTXOs = append(unUTXOs, utxo)
					}

				} else {
					//还没有花费记录,直接向未花费直接添加
					utxo := &UTXO{tx.TxHash, index, out}
					unUTXOs = append(unUTXOs, utxo)
				}

			}

		}
	}

	//--------在次查询已经在数据库存储的交易--------------
	//通过迭代器去遍历区块
	blockchainIterator := blockchain.Iterator()
	for {
		block := blockchainIterator.Next()
		fmt.Println(block)

		//遍历每个block的交易
		//for _, tx := range block.Txs {
		//倒叙遍历交易列表,否则每个区块中有多条数据时数据统计有误
		for i := len(block.Txs) - 1; i >= 0; i-- {
			tx := block.Txs[i]
			//txHash

			//Vins
			//保证他不是coinbase交易   //主要提高效率,直接忽略遍历coinbase交易
			if !tx.IsCoinbaseTransaction() {
				for _, in := range tx.Vins {
					//查看已经花费的,是否已经解锁

					version_pubKey_checksum := Base58Decode([]byte(address))
					pubKeyHash := version_pubKey_checksum[1:len(version_pubKey_checksum)-4]

					if in.UnlockRipemd160(pubKeyHash) {
						//十六进制字符串
						key := hex.EncodeToString(in.TXHash)
						//主要是相同的交易hash可能存在多条输出
						spentTXOutput[key] = append(spentTXOutput[key], in.Vout)
					}
				}
			}

			//Vouts
			//遍历出来  包含已经花费和未花费
			//定义标签名字:
		outputs:
			for index, out := range tx.Vouts {
				if out.UnlockScriptPubKeyWithAddress(address) {
					//如果已花费的map不为空
					if len(spentTXOutput) != 0 {
						var isTxOutput bool //false   首先默认是未花费

						for txHash, indexArray := range spentTXOutput {

							for _, i := range indexArray {
								if index == i && txHash == hex.EncodeToString(tx.TxHash) {
									isTxOutput = true
									continue outputs
								}
							}
						}
						//将未花费的记录output添加到数组当中
						if !isTxOutput {
							utxo := &UTXO{tx.TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}

					} else {
						//还没有花费记录,直接向未花费直接添加
						utxo := &UTXO{tx.TxHash, index, out}
						unUTXOs = append(unUTXOs, utxo)
					}

				}

			}
		}

		//判断结束条件
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
	return unUTXOs
}

//转账时查找可用的UTXO    需要考虑还未打包的transaction
func (blockchain *Blockchain) FindSpendableUTXOS(from string, amount int, txs []*Transaction) (int64, map[string][]int) {
	//1.获取所有的UTXO
	utxos := blockchain.UnUTXOs(from, txs)

	//声明字典
	spendableUTXO := make(map[string][]int)
	//2.遍历utxos
	var value int64
	for _, utxo := range utxos {

		value = value + utxo.Output.Value

		//hash转换string
		txHash := hex.EncodeToString(utxo.TXHash)
		spendableUTXO[txHash] = append(spendableUTXO[txHash], utxo.Index)

		if value >= int64(amount) {
			break
		}
	}
	//余额不足
	if value < int64(amount) {
		fmt.Printf("%s's的余额不足", from)
		os.Exit(1)
	}
	return value, spendableUTXO
}

//挖掘新的区块
func (blockchain *Blockchain) MineNewBlock(from []string, to []string, amount []string, nodeID string) {

	//1.建立一笔交易
	fmt.Println(from)
	fmt.Println(to)
	fmt.Println(amount)

	utxoSet := &UTXOSet{blockchain}
	//1.通过相关算法建立transaction数组
	var txs []*Transaction

	for index, address := range from {
		value, _ := strconv.Atoi(amount[index])

		//要把本次的交易数据进行考虑
		tx := NewSimpleTransaction(address, to[index], int64(value), utxoSet, txs, nodeID)

		txs = append(txs, tx)
	}

	//奖励机制
	tx := NewCoinBaseTransaction(from[0])
	txs = append(txs, tx)

	var block *Block
	blockchain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlockBucket))
		if b != nil {
			hash := b.Get([]byte("l"))
			blockBytes := b.Get(hash)
			block = DeserializeBlock(blockBytes)
		}
		return nil
	})

	//建立新的区块之前进行交易数据的验证(对tx进行签名验证)
	for _, tx := range txs {
		if blockchain.VerifyTransaction(tx, txs) == false {
			log.Panic("-------验签失败-----")
		}
	}

	//2.建立新的区块
	block = NewBlock(txs, block.Height+1, block.Hash)

	//将新区快存储到数据库
	blockchain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlockBucket))
		if b != nil {
			b.Put(block.Hash, block.Serialize())

			b.Put([]byte("l"), block.Hash)

			blockchain.Tip = block.Hash
		}
		return nil
	})
}

//查询余额
func (blockchain *Blockchain) GetBalance(address string) int64 {
	//对于查询余额,只需要传空就可以了
	utxos := blockchain.UnUTXOs(address, []*Transaction{})
	var amount int64
	//遍历所有的未花费的余额
	for _, utxo := range utxos {
		amount = amount + utxo.Output.Value
	}
	//查询余额
	return amount
}

//签名交易
func (blockchain *Blockchain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey, txs []*Transaction) {
	//1.coinbase交易是不需要进行签名验证的
	if tx.IsCoinbaseTransaction() {
		return
	}
	//2.获取该tx中的Input,引用之前的transaction中的未花费的output
	//找到Vins里面每个交易Hash所对应的交易列表
	prevTXs := make(map[string]*Transaction)

	for _, vin := range tx.Vins {
		//--------------------------注意,根据vin的ID,不仅仅要查找上了区块的交易,未打包上链的交易----------------
		prevTX, err := blockchain.FindTransaction(vin.TXHash, txs)
		if err != nil {
			log.Panic(err)
		}
		//根据Transaction的交易hash,保存transaction对象
		prevTXs[hex.EncodeToString(prevTX.TxHash)] = &prevTX
	}

	//进行签名  签名需要私钥
	tx.Sign(privateKey, prevTXs)
}

//根据交易iD获取所对应的交易对象          遍历数据库查找的,切记要记着还未上块的交易列表
func (blockchain *Blockchain) FindTransaction(ID []byte, txs []*Transaction) (Transaction, error) {
	//1.先检查未打包的的txs
	for _, tx := range txs {
		if bytes.Compare(ID, tx.TxHash) == 0 {
			return *tx, nil
		}
	}

	//2.遍历数据库,获取block--->transaction
	iterator := blockchain.Iterator()
	for {
		block := iterator.Next()

		for _, tx := range block.Txs {
			if bytes.Compare(tx.TxHash, ID) == 0 {
				return *tx, nil
			}
		}

		//判断循环结束的条件
		var hashInt *big.Int
		hashInt = big.NewInt(1)
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(hashInt) == 0 {
			break
		}
	}
	return Transaction{}, nil
}

//验证数字签名
func (bloackchain *Blockchain) VerifyTransaction(tx *Transaction, txs []*Transaction) bool {
	//要想验证交易,首先获得之前的交易数据,之前的交易存入到map当中
	prevTXs := make(map[string]*Transaction)
	for _, vin := range tx.Vins {
		prevTX, err := bloackchain.FindTransaction(vin.TXHash, txs)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.TxHash)] = &prevTX
	}
	return tx.Verify(prevTXs)
}

//返回未花费所存储的map
func (blockchain *Blockchain) FindUTXOMap() map[string]*TXOutputs {
	iterator := blockchain.Iterator()
	//定义一个字典Map,用于存储未花费的输出以及交易hash

	//(1)首先记录已经花费的所有的Input,主要是存储已经花费的交易信息
	spentableUTXOsMap := make(map[string][]*TXInput)

	utxoMaps := make(map[string]*TXOutputs)

	for {
		//拿到最新的区块
		block := iterator.Next()
		//要倒叙遍历区块里面的信息

		for i := len(block.Txs) - 1; i >= 0; i-- {
			txOutputs := &TXOutputs{[]*UTXO{}}

			tx := block.Txs[i]

			//确定不是coinbase交易,进入---
			if tx.IsCoinbaseTransaction() == false {
				//遍历眉笔交易的input,将其内部的信息存入到已花费的map集合当中
				for _, txInput := range tx.Vins {
					txInputHash := hex.EncodeToString(txInput.TXHash)
					spentableUTXOsMap[txInputHash] = append(spentableUTXOsMap[txInputHash], txInput)
				}
			}
		WorkOutLoop:
			for index, out := range tx.Vouts {

				txOutputHash := hex.EncodeToString(tx.TxHash)
				txInputs := spentableUTXOsMap[txOutputHash]

				if len(txInputs) > 0 {
					var isSpent bool
					for _, input := range txInputs {

						outPubKey := out.Ripemd160Hash
						inPublickey := input.PublicKey

						if bytes.Compare(outPubKey, Ripemd160Hash(inPublickey)) == 0 {
							if index == input.Vout {
								isSpent = true
								continue WorkOutLoop
							}
						}
					}

					if !isSpent {
						//未花费
						utxo := &UTXO{tx.TxHash, index, out}
						txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
					}

				} else {
					//未花费
					utxo := &UTXO{tx.TxHash, index, out}
					txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
				}
			}
			//设置键值对
			//将当前的这个tx中,把 未花费的txoutputs ,存入到未花费的map当中
			//if len(txOutputs.UTXOS)>0{
			//	utxoMaps[hex.EncodeToString(tx.TxHash)]=txOutputs
			//}
			utxoMaps[hex.EncodeToString(tx.TxHash)] = txOutputs

		}
		//退出循环条件
		hashInt := big.NewInt(0)
		hashInt.SetBytes(block.PrevBlockHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return utxoMaps
}

//添加一个方法,用于查询最后一个区块的高度
func (blockchain *Blockchain) GetBestHeight() int64 {
	block := blockchain.Iterator().Next()
	return block.Height
}

//添加一个方法  查询当前区块链当中区块的hash值
func (blockchain *Blockchain) GetBlocksHashs() [][]byte {
	iterator := blockchain.Iterator()
	var hashes [][]byte
	for {
		block := iterator.Next()

		hashes = append(hashes, block.Hash)

		//判断迭代器的结束条件
		var hashInt *big.Int
		hashInt = big.NewInt(1)

		hashInt.SetBytes(block.PrevBlockHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return hashes
}

//添加一个方法,根据对应的hash值,来查找对应的区块
func (blockchain *Blockchain) GetBlockByHash(hash []byte) *Block {
	//(1)迭代器方法遍历
	//(2)直接查找数据库
	var block *Block
	err := blockchain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlockBucket))
		if b != nil {
			blockBytes := b.Get(hash)
			block = DeserializeBlock(blockBytes)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return block
}

//将传来的block对象存入到数据库当中
func (blockchian *Blockchain) AddBlock(block *Block) {

	//更新数据库
	//根据hash值,在数据库已经存在 不存储 直接结束 不存在  把区块数据存入到数据库
	err := blockchian.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlockBucket))
		if b != nil {
			//判断区块是否存在
			blockBytes := b.Get(block.Hash)
			if blockBytes != nil {
				return nil
			}
			//区块不存在,将区块存入到数据库当中
			err := b.Put(block.Hash, block.Serialize())
			if err != nil {
				log.Panic(err)
			}
			//更新blockchain当中最新的hash值
			//主要为了避免同步区块的异步问题,判断区块的高度,来进行设置区块的最新hash值所对应的 l
			blockHash := b.Get([]byte("l"))
			lastBlockBytes := b.Get(blockHash)
			//最后一个区块的最新hash值  根据区块高度来更新值
			lastBlock := DeserializeBlock(lastBlockBytes)
			if lastBlock.Height < block.Height {
				b.Put([]byte("l"), block.Hash)
				blockchian.Tip = block.Hash
			}

		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}
