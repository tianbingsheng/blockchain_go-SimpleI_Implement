 package BLC

import (
	"fmt"
	"os"
	"flag"
	"log"
)

type CLI struct {
}

func PrintUsage() {
	fmt.Println("Usage")
	fmt.Println("\tcreatewallet --创建钱包")
	fmt.Println("\taddresslists --输出所有钱包地址")
	fmt.Println("\tcreateblockchain -address --创建创世区块的地址信息")
	fmt.Println("\tsend -from FROM -to TO -amount AMOUNT --交易明细")
	fmt.Println("\tprintchain --输出区块信息")
	fmt.Println("\tgetbalance -address --钱包余额")
	fmt.Println("\ttest --测试")
	fmt.Println("\tstartnode -miner Address --启动节点并指定挖矿的奖励地址")
}

func IsValidArgs() {
	if len(os.Args) < 2 {
		PrintUsage()
		os.Exit(1)
	}
}

//获取终端窗口配置的环境变量

func (cli *CLI) Run() {
	IsValidArgs()

	/*
变量名 :NODE_ID=9527
获取终端所配置的node_id的值
os.Getenv(变量名)---->变量值
*/

	node_id := os.Getenv("NODE_ID")
	if node_id == "" {
		fmt.Println("没有设置node_id,程序即将结束")
		os.Exit(1)
	}
	fmt.Println("NODE_ID",node_id)


	testCmd := flag.NewFlagSet("test", flag.ExitOnError)
	addresslistsCmd := flag.NewFlagSet("addresslists", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	sendBlockCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getbalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	startnodeCmd := flag.NewFlagSet("startnode",flag.ExitOnError)



	flagFrom := sendBlockCmd.String("from", "", "转账源地址")
	flagTo := sendBlockCmd.String("to", "", "转账目的地地址")
	flagAmount := sendBlockCmd.String("amount", "", "转账金额")
	getBalanceWithAddress := getbalanceCmd.String("address", "", "要查询某一个账号的余额")
	flagcreateBlockchainWithAddress := createBlockchainCmd.String("address", "", "创建创世区块的地址")
	flagMinerData := startnodeCmd.String("miner","","启动节点并指定挖矿的奖励地址")

	//解析数据
	switch os.Args[1] {
	case "send":
		err := sendBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getbalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "addresslists":
		err := addresslistsCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "test":
		err := testCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startnodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		PrintUsage()
		os.Exit(1)
	}

	//判断是否已经解析成功
	if sendBlockCmd.Parsed() {
		//如果解析数据是空,打印信息直接退出
		if *flagFrom == "" || *flagTo == "" || *flagAmount == "" {
			PrintUsage()
			os.Exit(1)
		}
		//指定标准的输入格式,正则表达式指定指定的格式
		//三个数组的个数是相等的

		//打印解析后的数据
		//fmt.Println(*flagAddBlockData)
		//cli.addBlock([]*Transaction{})
		//fmt.Println(*flagFrom)
		//fmt.Println(*flagTo)
		//fmt.Println(*flagAmount)
		//转换后的[]string
		//fmt.Println(JSONTOArray(*flagFrom))
		//fmt.Println(JSONTOArray(*flagTo))
		//fmt.Println(JSONTOArray(*flagAmount))

		from := JSONTOArray(*flagFrom)
		to := JSONTOArray(*flagTo)
		//验证转账钱包地址的合法有效性
		for index, fromAddress := range from {
			if IsValidForAddress([]byte(fromAddress)) == false || IsValidForAddress([]byte(to[index])) == false {
				fmt.Println("-----------地址无效,不合法---------")
				PrintUsage()
				os.Exit(1)
			}
		}
		//for _,toAddress := range to {
		//	IsValidForAddress([]byte(toAddress))
		//}
		amount := JSONTOArray(*flagAmount)
		cli.Send(from, to, amount,node_id)

	}

	if printChainCmd.Parsed() {
		//fmt.Println("输出所有区块的数据..........")
		cli.Printchain(node_id)
	}
	//创建创世区块
	if createBlockchainCmd.Parsed() {

		if *flagcreateBlockchainWithAddress == "" {
			fmt.Println("地址不能为空......")
			PrintUsage()
			os.Exit(1)
		}

		//验证地址的合法有效性
		if !IsValidForAddress([]byte(*flagcreateBlockchainWithAddress)) {
			fmt.Println("---------地址不合法-------")
			PrintUsage()
			os.Exit(1)
		}
		cli.createGenesisBlockchain(*flagcreateBlockchainWithAddress,node_id)
	}
	//查询余额
	if getbalanceCmd.Parsed() {

		if *getBalanceWithAddress == "" {
			fmt.Println("地址不能为空")
			PrintUsage()
			os.Exit(1) //0 :正常退出   1 :非正常退出
		}
		if !IsValidForAddress([]byte(*getBalanceWithAddress)) {
			fmt.Println("------地址不合法-----")
			PrintUsage()
			os.Exit(1)
		}
		//根据钱包地址查询余额
		cli.getBalance(*getBalanceWithAddress,node_id)
	}
	//创建钱包
	if createWalletCmd.Parsed() {
		cli.createWallet(node_id)
	}
	//钱包地址列表
	if addresslistsCmd.Parsed() {
		cli.addressLists(node_id)
	}
	if testCmd.Parsed() {
		cli.TestMethod(node_id)
	}
	if startnodeCmd.Parsed(){

		cli.StartNode(node_id,*flagMinerData)
	}

}
