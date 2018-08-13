package BLC

import (
	"fmt"
	"os"
)

//输出打印区块链
func (cli *CLI) Printchain(nodeID string) {
	DBName := fmt.Sprintf(dbName,nodeID)
	if DBExists(DBName) == false {
		fmt.Println("数据库不存在-----")
		os.Exit(1)
	}
	blockchain := BlockchainObject(nodeID)
	defer blockchain.DB.Close()
	blockchain.Printchain()
}