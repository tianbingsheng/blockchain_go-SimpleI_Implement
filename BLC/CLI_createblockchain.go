package BLC

//创建创世区块
func (cli *CLI) createGenesisBlockchain(address string,node_ID string) {
	//1.创建一个coinBase交易

	blockchain := CreateBlockchainWithGenesis(address,node_ID)
	//地址传递  数据库只能关闭一次
	defer blockchain.DB.Close()
	//注意必须传地址
	utxoSet := &UTXOSet{blockchain}
	utxoSet.ResetUTXOSet()
	//其实已经在数据库当中存储了两张表
}