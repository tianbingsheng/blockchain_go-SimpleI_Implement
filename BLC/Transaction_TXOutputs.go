package BLC

import (
	"bytes"
	"encoding/gob"
	"log"
)

//存储未花费的输出
type TXOutputs struct {
	//TxHash	  []byte
	UTXOS  []*UTXO
}

func (txOutputs *TXOutputs) Serialize() []byte {
	//创建缓冲区
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(txOutputs)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}
//反序列化
func DeserializeTXOutputs(txOutputsBytes []byte) *TXOutputs {
	var txOutputs TXOutputs
	decoder := gob.NewDecoder(bytes.NewReader(txOutputsBytes))
	err := decoder.Decode(&txOutputs)
	if err != nil {
		log.Fatal(err)
	}
	return &txOutputs
}