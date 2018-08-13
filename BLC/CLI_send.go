package BLC

import (
	"fmt"
	"os"
)

//转账
func (cli *CLI) Send(from []string, to []string, amount []string,nodeID string) {
	DBName := fmt.Sprintf(dbName,nodeID)
	if DBExists(DBName) == false {
		fmt.Println("-------数据不存在------")
		os.Exit(1)
	}
	blockchain := BlockchainObject(nodeID)
	defer blockchain.DB.Close()

	blockchain.MineNewBlock(from, to, amount,nodeID)
	utxoSet := &UTXOSet{blockchain}
	//转账成功以后,需要更新数据库
	utxoSet.Update()
}