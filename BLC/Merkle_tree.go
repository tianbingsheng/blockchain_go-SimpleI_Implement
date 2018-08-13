package BLC

import (
	"crypto/sha256"
	"math"
)

//1.Merkle Tree   创建结构体对象 表示节点和树
type MerkleTree struct {
	RootNode *MerkleNode
}

//Merkle Node
type MerkleNode struct {
	LeftNode  *MerkleNode
	RightNode *MerkleNode
	DataHash  []byte
}

//2.给一个左右节点  生成一个新的节点
//叶子节点   非叶子节点
func NewMerkleNode(leftNode, rightNode *MerkleNode, txHash []byte) *MerkleNode {
	//创建当前节点
	mNode := &MerkleNode{}

	//2.赋值
	if leftNode == nil && rightNode == nil {
		//mNode就是叶子节点
		hash := sha256.Sum256(txHash)
		mNode.DataHash = hash[:]
	} else {
		//mNode是非叶子节点
		prevHash := append(leftNode.DataHash, rightNode.DataHash...)
		hash := sha256.Sum256(prevHash)
		mNode.DataHash = hash[:]
	}
	mNode.LeftNode = leftNode
	mNode.RightNode = rightNode

	return mNode
}

//3.Generate Merkle RootNood
func NewMerkleTree(txHashData [][]byte) *MerkleTree {
	//Tx1 Tx2 Tx3
	//1.首先创造一个存Merkle节点的数组
	var nodes []*MerkleNode
	//2.首先判断当前交易的奇偶性
	if len(txHashData)%2 != 0 {
		//基数,复制最后一个
		txHashData = append(txHashData, txHashData[len(txHashData)-1])
	}
	//3创建一排的叶子节点
	for _, datum := range txHashData {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, node)
	}
	//思路直接for true 判断其个数  (2)幂的结构
	//4.生成根节点
	count:= GetCircleCount(len(nodes))
	for i := 0; i < count; i++ {
		var newLevel []*MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(nodes[j], nodes[j+1], nil)
			newLevel = append(newLevel, node)
		}
		//if len(nodes) == 1{
		//	break
		//}
		//判断Newlevel的奇偶性
		if len(newLevel)%2!= 0{
			newLevel = append(newLevel,newLevel[len(newLevel)-1])
		}

		nodes = newLevel
	}

	mTree := &MerkleTree{nodes[0]}
	return mTree
}
//确定循环次数
func GetCircleCount(len int)int{
	count := 0
	for {
		if int(math.Pow(2,float64(count)))> len {
			return count
		}
		count++
	}
}





















