package BLC

import "fmt"

func (cli *CLI)createWallet(nodeID string){
	wallets ,_:= NewWallets(nodeID)
	wallets.CreateNewWallet(nodeID)

	fmt.Println(len(wallets.WalletsMap))
}
//单例模式的声明周期只是在整个程序的运行过程当中
//本次案例只能创建本地文件,读取写入数据
