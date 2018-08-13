package BLC

import (
	"math/big"
	"bytes"
	"crypto/sha256"
	"fmt"
)

//0000 0000 0000 0000 1001 0001 0000 .... 0001

//256位hash里面前面至少有16个0
const targetBit = 16

type ProofOfwork struct {
	Block *Block		//当前需要验证的区块
	Target *big.Int 	//大数存储,区块难度  数字很大,不会溢出 代表难度
}
//数据拼接,返回字节数组
func (pow *ProofOfwork)PrepareData(nouce int64)[]byte{
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevBlockHash,
			pow.Block.HashTransactions(),
			IntToHex(int64(nouce)),
			IntToHex(pow.Block.Timestamp),
			IntToHex(int64(targetBit)),	//加不加都可以
			IntToHex(pow.Block.Height),
		},[]byte{},
	)
	return data
}
func(proofOfwork *ProofOfwork) Run()([]byte,int64){
	//1.将Block的属性拼接成字节数组

	//2.生成hash
	//3.判断hash的有效性,如果满足条件,跳出循环
	var hashInt *big.Int		//存储我们新生成的hash值
	hashInt = big.NewInt(0)	//进行简单的实例化操作,否则对象为nil

	var nouce int64 = 0
	var hash [32]byte

	for {
		//准备数据
		dataBytes := proofOfwork.PrepareData(nouce)
		//生成hash值
		hash = sha256.Sum256(dataBytes)
		fmt.Printf("\r%x",hash)
		//将hash存储到hashInt
		hashInt.SetBytes(hash[:])
		//判断hashInt是否小于block里面的target
		if proofOfwork.Target.Cmp(hashInt) == 1{
			break
		}
		nouce++
	}
	return hash[:],nouce
}
//返回一个工作量证明的对象,创建一个新的工作量证明那个对象
func NewProofOfWork (block *Block)*ProofOfwork{

	//1.big.Int对象 1
	//0000 0001
	//8-2 = 6
	//0100 0000

	//1.创建一个初始值为1的target
	target := big.NewInt(1)
	//2.左移256 - targetBit
	target.Lsh(target,256-targetBit)

	return &ProofOfwork{block,target}
}






















