package BLC

const dbName = "blockchain_%s.db" //数据库的名字
/*
数据库名字:
blockchain_port.db
blockchain_3000.db
blockchain_3001.db
*/
const BlockBucket = "blocks" //桶的名字

//定义当前成许的版本
const NODE_VERSION = 1

//定义命令的长度
const COMMANDLENGTH = 12

//具体的命令
//发送版本
const COMMAND_VERSION = "version"

const COMMAND_GETDATA = "getdata"

const COMMAND_BLOCKDATA  = "blockdata"


//发送blocks

const COMMAND_GETBLOCKS = "getblocks"

//定以类型
//区块类型
const BLOCK_TYPE = "block"

//交易类型
const TX_TYPE = "tx"

const COMMAND_INV = "inv"