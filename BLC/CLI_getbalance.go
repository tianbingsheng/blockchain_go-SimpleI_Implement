package BLC

import "fmt"

//先用他去查询余额
func (cli *CLI)getBalance(address string,nodeID string){

	fmt.Println(address)
	blockchain := BlockchainObject(nodeID)
	defer blockchain.DB.Close()
	utxoSet := &UTXOSet{blockchain}
	//amount := blockchain.GetBalance(address)
	//查询余额代码优化,不是遍历整个数据库区块
	amount := utxoSet.GetBalance(address)
	fmt.Printf("用户%s的余额总共是%d枚BTC\n",address,amount)
}
//余额查询优化
//之前的查询余额会遍历整个数据库去查找所有的区块  这样效率会很低
