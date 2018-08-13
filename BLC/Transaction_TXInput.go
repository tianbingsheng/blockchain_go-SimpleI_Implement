package BLC

import "bytes"

//交易输入
type TXInput struct {
	TXHash []byte //交易的Hash值
	Vout   int    //存储TXOutput在Vout里面的索引
	Signature []byte	//数字签名
	PublicKey    []byte	//公钥  钱包里面的

	//ScriptSig string //数字签名    (理解为用户名)
}

//解锁你要花费哪笔钱的

//判断当前的消费是谁的钱
//input                                    对address进行解码后的pubkeyHash
func (txInput *TXInput) UnlockRipemd160(ripemd160Hash []byte) bool {

	return bytes.Compare(Ripemd160Hash(txInput.PublicKey),ripemd160Hash) == 0
}
