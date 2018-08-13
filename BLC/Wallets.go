package BLC

import (
	"fmt"
	"bytes"
	"encoding/gob"
	"log"
	"io/ioutil"
	"crypto/elliptic"
	"os"
)

//定义多个钱包的结构体对象
type Wallets struct {
	WalletsMap map[string]*Wallet
}

//格式化钱包的文件名字
const WalletFile = "wallet_%s.dat"

//创建钱包的集合
func NewWallets(nodeID string) (*Wallets, error) {
	//首先判断钱包文件在不在
	//如果不存在
	walletFile := fmt.Sprintf(WalletFile, nodeID)

	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		wallets := &Wallets{}
		wallets.WalletsMap = make(map[string]*Wallet)
		return wallets, err
	}
	//本地的钱包信息取出来
	wallets := &Wallets{}
	wallets.LoadFromFile(nodeID) //加载数据库文件
	return wallets, nil
}

//创建一个新的钱包
func (w *Wallets) CreateNewWallet(nodeID string) {
	wallet := NewWallet()
	fmt.Printf("Address :%s\n", wallet.GetAddress())
	//创建完成新的钱包之后,保存到wallets钱包集合当中
	w.WalletsMap[string(wallet.GetAddress())] = wallet
	//添加钱包数据后,在把所有wallets数据保存到文件系统中
	w.SaveWallets(nodeID)
}

//将钱包信息写入到数据库文件当中
func (w *Wallets) SaveWallets(nodeID string) {
	//首先将wallets序列化
	var content bytes.Buffer

	//注册的目的,是为了可以序列化任何类型,(包括接口) wallets结构复杂(包括各种曲线)
	////序列化的过程中：被序列化的对象 中包含了接口，那么将口需要注册
	//比如一个结构体中包含一个接口类型的内容,序列化时就必须进行注册服务
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(w)
	if err != nil {
		log.Panic(err)
	}
	//序列化后的数据 就是 content.Bytes()
	//将序列化后的数据写入到文件当中
	walletFile := fmt.Sprintf(WalletFile, nodeID)
	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

//加载钱包文件
func (w *Wallets) LoadFromFile(nodeID string) error {
	walletFile := fmt.Sprintf(WalletFile, nodeID)
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}
	//读取文件内容
	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var wallets Wallets
	//序列化时注册,在反序列化时也要进行一次注册服务
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}
	w.WalletsMap = wallets.WalletsMap
	return nil
}
