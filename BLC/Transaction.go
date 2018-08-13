package BLC

import (
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
	"encoding/hex"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/elliptic"
	"math/big"
	"time"
)

//UTXO未花费的交易输出
type Transaction struct {
	//1.交易hash(每一笔交易都有自己的hash值)
	TxHash []byte

	//2.输入
	Vins []*TXInput

	//3.输出
	Vouts []*TXOutput
}

//创世区块的交易信息数据比较特殊
//创世区块的Vins   Vouts 里面只有一个对象\
//Transaction 创建分两种情况

//判断当前的交易是否是coinbase交易
func (tx *Transaction)IsCoinbaseTransaction()bool{
	return len(tx.Vins[0].TXHash)==0 && tx.Vins[0].Vout == -1
}



//1.创世区块创建时transaction         创建创世区块的时候,创世区块里面的第一笔交易就是coinbase交易
func NewCoinBaseTransaction(address string) *Transaction {
	//代表消费
	txInput := &TXInput{[]byte{}, -1, nil,[]byte{}}
	// 转账 找零
	txOutput := NewTXOutput(10,address)

	//txOutput := &TXOutput{10, address}

	txCoinbase := &Transaction{[]byte{}, []*TXInput{txInput}, []*TXOutput{txOutput}}
	//设置hash值
	txCoinbase.HashTransaction()

	return txCoinbase
}

//
func (tx *Transaction)HashTransaction() {
	//创建缓冲区
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	resultBytes := bytes.Join([][]byte{IntToHex(time.Now().Unix()),result.Bytes()},[]byte{})
	hash := sha256.Sum256(resultBytes)
	tx.TxHash = hash[:]
}

// 2.转账的时候产生的交易   建立交易 返回对应的交易 普通的transaction
func NewSimpleTransaction(from string, to string, amount int64,utxoSet *UTXOSet,txs []*Transaction,nodeID string) *Transaction {
	//$ ./main send -from '["tianbingsheng"]' -to '["qiuxinhua"]' -amount '["4"]'   10  4
	//           qiuxinhua   -to zhangsan  2                  4 -  2
	//			tianbingsheng  -to zhangsan                   6 -  2


	//有一个函数,返回from这个人所有的为花费交易输出所对应的Transaction 在一个transaction当中一个人的名下可能会有多个为花费的交易输出,假如找零分批次找零
	//获得持有人所有的未花费的UTXOS模型
	//通过钱包获取公钥的思路

	wallets ,_:= NewWallets(nodeID)
	wallet := wallets.WalletsMap[from]

	//通过一个函数,返回可用金额 以及字典map
	money,spendableUTXODic := utxoSet.FindSpendableUTXOS(from,amount,txs)

	var txInputs []*TXInput
	var txOutputs []*TXOutput

	//代表消费 input
	for txHash,indexArray:= range spendableUTXODic{
		for _,index := range indexArray{
			txHashBytes,_:=hex.DecodeString(txHash)
			txInput := &TXInput{txHashBytes, index, nil,wallet.PublicKey}
			txInputs = append(txInputs, txInput)
		}
	}


	// 转账
	//txOutput := &TXOutput{int64(amount), to}
	txOutput := NewTXOutput(int64(amount),to)
	txOutputs = append(txOutputs, txOutput)

	//找零
	txOutput = NewTXOutput(int64(money)-int64(amount),from)
	//txOutput = &TXOutput{int64(money) - int64(amount), from}
	txOutputs = append(txOutputs, txOutput)

	//交易列表填充
	tx := &Transaction{[]byte{}, txInputs, txOutputs}
	//设置hash值
	tx.HashTransaction()

	//进行数字签名
	utxoSet.Blockchain.SignTransaction(tx,wallet.PrivateKey,txs)
	return tx
}
//对Transaction中的每一个input进行签名
//签名 :为了对一笔交易进行签名
//私钥
//要获取交易的input 引用之前的output 所在之前的交易
func (tx *Transaction)Sign(privateKey ecdsa.PrivateKey ,prevTXs map[string]*Transaction){
	//判断当前交易是否是coinbase交易 coinbase交易不需要进行 签名验签
	if tx.IsCoinbaseTransaction(){
		return
	}
	//当前的input没有找到Transaction
	//获取input对应的output所在的tx,如果不存在,无法进行签名
	for _,vin := range tx.Vins{
		if prevTXs[hex.EncodeToString(vin.TXHash)] == nil {
			log.Panic("当前的Input,没有找到对应的output所在的transaction,无法进行签名")
		}
	}

	//交易副本 即将进行签名,要签名的数据
	txCopy := tx.TrimmedCopy()

	for inID,vin := range txCopy.Vins{
		prevtx := prevTXs[hex.EncodeToString(vin.TXHash)]
		//为了保险,又重新置空,仅仅是为了保证签名一定为空
		txCopy.Vins[inID].Signature = nil
		//为了验证,存储的是公钥hash,不是原始公钥
		//设置input中publicKey为对应的output的pubKeyHash
		txCopy.Vins[inID].PublicKey = prevtx.Vouts[vin.Vout].Ripemd160Hash
		txCopy.TxHash = txCopy.Hash()
		//为了方便下一个input,主要为接下来下一轮迭代,生成hash条件一致
		txCopy.Vins[inID].PublicKey=nil

		//签名代码
		//第一个参数就是随机值  第二个参数:私钥  第三个参数:要签名的数据(签署的是修剪后的交易副本)
		r,s,err := ecdsa.Sign(rand.Reader,&privateKey,txCopy.TxHash)
		if err != nil {
			log.Panic(err)
		}
		//获得签名
		signature := append(r.Bytes(),s.Bytes()...)
		//获得签名,给签名字段赋值签名
		tx.Vins[inID].Signature = signature
	}

}
//拷贝一份新的Transaction用于填写签名
//要签名的tx中,并不是所有的数据都要作为签名数据,生成签名
//只需要签名所需要的部分
/*
交易的副本中包含的数据:
包含了 tx 中的输出和输入
	inputs : sign publicKey 不要
	outputs :
*/
func (tx *Transaction)TrimmedCopy()Transaction{
	var inputs []*TXInput
	var outputs []*TXOutput

	for _,vin := range tx.Vins{
		inputs = append(inputs,&TXInput{vin.TXHash,vin.Vout,nil,nil})
	}
	for _,out := range tx.Vouts{
		outputs = append(outputs,&TXOutput{out.Value,out.Ripemd160Hash})
	}
	txCopy := Transaction{tx.TxHash,inputs,outputs}
	return txCopy
}
func (tx *Transaction)Hash()[]byte{
	txCopy := tx
	txCopy.TxHash = []byte{}
	hash := sha256.Sum256(txCopy.Serialize())
	return hash[:]
}
func (tx*Transaction) Serialize() []byte {
	//创建缓冲区
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}
//实现数字签名的验签工作
func(tx *Transaction)Verify(prevTXs map[string]*Transaction)bool{
	//其实coinbase交易不存在签名验签过程
	if tx.IsCoinbaseTransaction(){
		return true
	}
	for _,vin := range tx.Vins{
		if prevTXs[hex.EncodeToString(vin.TXHash)] == nil {
			log.Panic("根据input没有找到output所对应的Transaction交易")
		}
	}
	txCopy := tx.TrimmedCopy()

	curve := elliptic.P256()	//获得椭圆曲线

	for inID,vin := range tx.Vins{
		prevTX := prevTXs[hex.EncodeToString(vin.TXHash)]
		txCopy.Vins[inID].Signature = nil
		txCopy.Vins[inID].PublicKey = prevTX.Vouts[vin.Vout].Ripemd160Hash
		txCopy.TxHash = txCopy.Hash()  //要签名的数据
		txCopy.Vins[inID].PublicKey = nil


		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen/2)])
		s.SetBytes(vin.Signature[(sigLen/2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len (vin.PublicKey)
		x.SetBytes(vin.PublicKey[:(keyLen/2)])
		y.SetBytes(vin.PublicKey[(keyLen/2):])

		rawPubKey := ecdsa.PublicKey{curve,&x,&y}
		//验证的是要需要交易的hash值
		//第一个参数:公钥
		if ecdsa.Verify(&rawPubKey,txCopy.TxHash,&r,&s) == false {
			return false
		}
	}
	return true
}

















