package BLC

import "bytes"

//交易输出
type TXOutput struct {
	Value         int64
	Ripemd160Hash []byte //公钥Hash
	//ScriptPubKey string //公钥  (理解为用户名)
}

//其实包含转账 找零
//解析地址,上锁,实质上是设置pubKeyHash值
func (out *TXOutput)Lock(address string){
	//拿到的字节数是25个字节
	version_pubKey_checksum:= Base58Decode([]byte(address))
	pubKeyHash := version_pubKey_checksum[1:len(version_pubKey_checksum)-4]
	out.Ripemd160Hash = pubKeyHash
}

func (txOutput *TXOutput) UnlockScriptPubKeyWithAddress(address string) bool {
	version_pubKey_checksum:= Base58Decode([]byte(address))
	pubKeyHash := version_pubKey_checksum[1:len(version_pubKey_checksum)-4]
	return bytes.Compare(txOutput.Ripemd160Hash,pubKeyHash) == 0
}
func NewTXOutput (value int64,address string)*TXOutput{
	txOutput := &TXOutput{value,nil}
	//设置Ripemd160Hash
	txOutput.Lock(address)
	return txOutput
}

