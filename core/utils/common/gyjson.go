package common

import (
	"encoding/json"
	"os"
)

// 读入文件，将文件中的json转为结构体
func ReadJsonFile(filename string, v interface{}) error {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	return json.NewDecoder(jsonFile).Decode(v)
}

// 将结构体写入到json文件
func WriteJsonFile(filename string, v interface{}) error {
	jsonFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	return json.NewEncoder(jsonFile).Encode(v)
}
