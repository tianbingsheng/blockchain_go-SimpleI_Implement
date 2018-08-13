package BLC

import (
	"bytes"
	"encoding/binary"
	"log"
	"encoding/json"
	"fmt"
	"encoding/gob"
)

//将int64转换为字节数组
func IntToHex (num int64) []byte {
	buff := new (bytes.Buffer)
	err := binary.Write(buff,binary.BigEndian,num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}
//标准的JSONTOArray
func JSONTOArray(jsonString string)[]string{
	//json到[]string
	var sArr []string
	err := json.Unmarshal([]byte(jsonString),&sArr)
	if err != nil {
		log.Panic(err)
	}
	return sArr
}
// 字节数组反转
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

//命令转换为字节,将给定的字符串命令转换为字节数组
func commandToBytes(command string) []byte {
	var bytes [COMMANDLENGTH]byte
	for i, c := range command {
		bytes[i] = byte(c)
	}
	return bytes[:]
}

//字节转换为命令,将给定的字节数组转换为string类型的命令
func BytesToCommand(bytes []byte)string{
	var command []byte
	for _,b := range bytes{
		if b != 0x0 {
			command = append(command,b)
		}
	}
	return fmt.Sprintf("%s",command)
}

//将对象进行序列化
func gobEncode(version interface{})[]byte{
	//创建缓冲区
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(version)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}















