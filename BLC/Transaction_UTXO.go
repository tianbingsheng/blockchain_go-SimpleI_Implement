package BLC

//未花费输出的结构对象
type UTXO struct {
	TXHash []byte
	Index  int
	Output  *TXOutput
}
//制造这样一个对象主要是为了能够记录在转账时所需要的交易参数信息
