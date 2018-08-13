package BLC

import (
	"github.com/boltdb/bolt"
	"log"
	"encoding/hex"
	"fmt"
	"bytes"
)

//1.有一个方法,功能:
//遍历整个数据库,读取所有的未花费的UTXO,然后将所有的UTXO存储到数据库当中
//就是说数据库当中有两张表   一个存储区块block信息 一个存储UTXO未花费的信息
//reset    就是要遍历整个数据库 花费的时间可能会很长
/*
(1)这是reset的相关操作原理:
去遍历数据库时, 返回一个map[string]*TXOutputs
返回[]*TXOutputs

(2)转账的时候,我直接从UTXO中查找相关数据即可
   当消费完成以后,要更新未花费输出的数据表(不更新简单粗暴就是重置数据库)  重置数据库会非常耗时

blocks:
utxoTable:

*/

const utxoTableName = "utxoTableName"

type UTXOSet struct {
	Blockchain *Blockchain
}

//重置数据库表    (遍历整个数据库)
func (utxoSet *UTXOSet) ResetUTXOSet() {
	fmt.Println("-----------1111111111---------")
	err := utxoSet.Blockchain.DB.Update(func(tx *bolt.Tx) error {
		fmt.Println("------------22222------------")
		//查看表是否存在
		b := tx.Bucket([]byte(utxoTableName))

		if b != nil {
			//删除表
			err := tx.DeleteBucket([]byte(utxoTableName))
			if err != nil {
				log.Panic(err)
			}
		}

		b, _ = tx.CreateBucket([]byte(utxoTableName))
		if b != nil {
			//[string]*TXOutputs   存入map当中
			txOutputsMap := utxoSet.Blockchain.FindUTXOMap()

			for keyHash, outs := range txOutputsMap {
				txHash, _ := hex.DecodeString(keyHash)
				b.Put(txHash, outs.Serialize())
			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}
func (utxoSet *UTXOSet) FindUTXOForAddress(address string) []*UTXO {
	//通过游标进行迭代遍历
	var utxos []*UTXO
	utxoSet.Blockchain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoTableName))
		//游标  可以不断遍历数据库所有的数据
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			txOutputs := DeserializeTXOutputs(v)
			for _, utxo := range txOutputs.UTXOS {
				if utxo.Output.UnlockScriptPubKeyWithAddress(address) {
					utxos = append(utxos, utxo)
				}
			}
		}
		return nil
	})
	return utxos
}

func (utxoSet *UTXOSet) GetBalance(address string) int64 {
	//1.先找到这个人名下的所有未花费的输出(未花费数据库)
	//2.然后在进行叠加
	utxos := utxoSet.FindUTXOForAddress(address)
	var amount int64
	for _, utxo := range utxos {
		amount = amount + utxo.Output.Value
	}
	return amount
}

//查询未打包的UTXOS
func (utxoSet *UTXOSet) FindUnpackSpendableUTXOS(from string, txs []*Transaction) ([]*UTXO) {

	var unUTXOs []*UTXO
	spentTXOutput := make(map[string][]int)

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
				version_pubKey_checksum := Base58Decode([]byte(from))
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
			if out.UnlockScriptPubKeyWithAddress(from) {
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
	return unUTXOs
}

//返回要凑多少钱,以及对应的TXOutput所对应的txHash以及相对应的index集合
func (utxoSet *UTXOSet) FindSpendableUTXOS(from string, amount int64, txs []*Transaction) (int64, map[string][]int) {
	unPackageUTXOS := utxoSet.FindUnpackSpendableUTXOS(from, txs)
	//首先遍历未打包交易的数据金额够不够,如果够直接跳过本地数据库
	var money int64 = 0
	spentableUTXOMap := make(map[string][]int)
	//遍历未打包的数据
	for _, utxo := range unPackageUTXOS {
		money = money + utxo.Output.Value

		TxHash := hex.EncodeToString(utxo.TXHash)
		spentableUTXOMap[TxHash] = append(spentableUTXOMap[TxHash], utxo.Index)

		if money >= amount {
			return money, spentableUTXOMap
		}
	}

	//钱还不够,遍历数据库的数据
	utxoSet.Blockchain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoTableName))
		if b != nil {
			c := b.Cursor()
		UTXOBREAK:
			for k, v := c.First(); k != nil; k, v = c.Next() {
				txOutputs := DeserializeTXOutputs(v)
				for _, utxo := range txOutputs.UTXOS {
					if utxo.Output.UnlockScriptPubKeyWithAddress(from) {
						money += utxo.Output.Value
						TxHash := hex.EncodeToString(utxo.TXHash)
						spentableUTXOMap[TxHash] = append(spentableUTXOMap[TxHash], utxo.Index)
						if money >= amount {
							break UTXOBREAK
						}
					}
				}
			}
		}
		return nil
	})

	if money < amount {
		log.Panic("------余额不足,请充值------")
	}
	return money, spentableUTXOMap
}

//更新数据库  打包好一个新区块后更新一次数据库
// 更新
func (utxoSet *UTXOSet) Update() {

	// 最新的Block
	block := utxoSet.Blockchain.Iterator().Next()

	ins := []*TXInput{}

	outsMap := make(map[string]*TXOutputs)

	// 找到所有我要删除的数据
	for _, tx := range block.Txs {

		for _, in := range tx.Vins {
			ins = append(ins, in)
		}
	}

	for _, tx := range block.Txs {

		utxos := []*UTXO{}

		for index, out := range tx.Vouts {

			isSpent := false

			for _, in := range ins {

				if in.Vout == index && bytes.Compare(tx.TxHash, in.TXHash) == 0 && bytes.Compare(out.Ripemd160Hash, Ripemd160Hash(in.PublicKey)) == 0 {

					isSpent = true
					continue
				}
			}

			if isSpent == false {
				utxo := &UTXO{tx.TxHash, index, out}
				utxos = append(utxos, utxo)
			}

		}

		if len(utxos) > 0 {
			txHash := hex.EncodeToString(tx.TxHash)
			outsMap[txHash] = &TXOutputs{utxos}
		}

	}

	err := utxoSet.Blockchain.DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(utxoTableName))

		if b != nil {

			// 删除
			for _, in := range ins {

				txOutputsBytes := b.Get(in.TXHash)

				if len(txOutputsBytes) == 0 {
					continue
				}

				fmt.Println("DeserializeTXOutputs")
				fmt.Println(txOutputsBytes)

				txOutputs := DeserializeTXOutputs(txOutputsBytes)

				fmt.Println(txOutputs)

				UTXOS := []*UTXO{}

				// 判断是否需要
				isNeedDelete := false

				for _, utxo := range txOutputs.UTXOS {

					if in.Vout == utxo.Index && bytes.Compare(utxo.Output.Ripemd160Hash, Ripemd160Hash(in.PublicKey)) == 0 {

						isNeedDelete = true
					} else {
						UTXOS = append(UTXOS, utxo)
					}
				}

				if isNeedDelete {
					b.Delete(in.TXHash)
					if len(UTXOS) > 0 {

						preTXOutputs := outsMap[hex.EncodeToString(in.TXHash)]

						preTXOutputs.UTXOS = append(preTXOutputs.UTXOS, UTXOS...)

						outsMap[hex.EncodeToString(in.TXHash)] = preTXOutputs

					}
				}

			}

			// 新增

			for keyHash, outPuts := range outsMap {
				keyHashBytes, _ := hex.DecodeString(keyHash)
				b.Put(keyHashBytes, outPuts.Serialize())
			}

		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

}
