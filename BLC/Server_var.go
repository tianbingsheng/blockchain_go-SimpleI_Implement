package BLC

var KonwNodes = []string{"localhost:3000"}

var nodeAddress string  //当前节点自己的地址

//存储还未同步的区块hash
var blocksArray [][]byte  //记录应该同步但是尚未同步的hash值