package BLC

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"bytes"
)

type Wallet struct {
	//1.私钥
	PrivateKey ecdsa.PrivateKey
	//2.公钥
	PublicKey []byte
}

const version = byte(0x00)
const addressChecksumLen = 4

//创建钱包
func NewWallet() *Wallet {
	//产生密钥对
	privateKey, publicKey := NewKeyPair()
	return &Wallet{PrivateKey: privateKey, PublicKey: publicKey}
}

//产生密钥对  通过私钥产生公钥
func NewKeyPair()(ecdsa.PrivateKey,[]byte){
	curve := elliptic.P256()  //椭圆曲线类型
	//私钥随机数无穷大 几乎随机不会重复(这样理解)
	privateKey ,err := ecdsa.GenerateKey(curve,rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	//私钥固定的话  相对的x和y的相对位置也是固定的
	publicKey := append(privateKey.PublicKey.X.Bytes(),privateKey.PublicKey.Y.Bytes()...)

	return *privateKey,publicKey
}

func (wallet *Wallet)GetAddress()[]byte{
	//1.先将publicKey进行一次sha256--->ripemd160
	ripemd160Hash:=Ripemd160Hash(wallet.PublicKey)

	//2.版本拼接
	version_ripemd160Hash := append([]byte{version},ripemd160Hash...)

	//3.进行两次的sha256运算
	checSumBytes := CheckSum(version_ripemd160Hash)

	//4.与版本,checksum进行拼接
	bytes := append(version_ripemd160Hash,checSumBytes...)

	//进行base58编码,并且返回 base58的编码长度会随着你编码长度的改变而改变  34字节固定长度输出(通常状况下是34)
	return Base58Encode(bytes)
}

func Ripemd160Hash(publicKey []byte)[]byte{
	//进行第一轮的sha256
	hash256 := sha256.New()
	hash256.Write(publicKey)
	hash := hash256.Sum(nil)

	//进行第二次的ripemd160
	ripemd160 := ripemd160.New()
	ripemd160.Write(hash)
	hash = ripemd160.Sum(nil)
	return hash
}
func CheckSum(payload []byte)[]byte{

	hash1 := sha256.Sum256(payload)
	hash2 := sha256.Sum256(hash1[:])
	//取出前四位
	return hash2[:addressChecksumLen]
}

//检测钱包地址的有效性

//在进行转账的时候,要判断钱包地址的有效合法性
func IsValidForAddress(address []byte)bool{
	version_public_checksumBytes := Base58Decode(address)

	checkBytes := version_public_checksumBytes[len(version_public_checksumBytes)-addressChecksumLen:]

	version_public := version_public_checksumBytes[:len(version_public_checksumBytes)-addressChecksumLen]

	check := CheckSum(version_public)
	//两个字节数组比较的方法
	if bytes.Compare(check,checkBytes) == 0 {
		return true
	}
	return false
}












