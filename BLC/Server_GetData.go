package BLC

//用于获取某个区块或者而某个交易的数据
type GetData struct {
	AddrFrom string //当前节点自己的地址
	Type     string //要获取的数据的类型
	Hash     []byte // 区块的的hash值  或者是tx交易的hash值
}
