package datatype

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// interface转字符串
func TypetoStr(data interface{}) string {
	return fmt.Sprintf("%v", data)
}

// interface转int
func TypetoInt(data interface{}) (int, error) {
	result := fmt.Sprintf("%v", data)
	return strconv.Atoi(result)
}

// interface转float64
func TypetoFloat64(data interface{}) (float64, error) {
	result := fmt.Sprintf("%v", data)
	return strconv.ParseFloat(result, 64)
}

// interface转bool
func TypetoBool(data interface{}) (bool, error) {
	result := fmt.Sprintf("%v", data)
	return strconv.ParseBool(result)
}

// interface转[]byte
func TypetoBytes(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// interface转map[string]interface{}
func TypetoMap(data interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := json.Unmarshal([]byte(TypetoStr(data)), &result)
	return result, err
}

// String 转  int
func StrtoInt(data string) (int, error) {
	return strconv.Atoi(data)
}

// String 转  float64
func StrtoFloat64(data string) (float64, error) {
	return strconv.ParseFloat(data, 64)
}

// String 转  bool
func StrtoBool(data string) (bool, error) {
	return strconv.ParseBool(data)
}

// String 转  []byte
func StrtoBytes(data string) ([]byte, error) {
	return json.Marshal(data)
}

// String 转  map[string]interface{}
func StrtoMap(data string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &result)
	return result, err
}

// String 转  []interface{}
func StrtoSlice(data string) ([]interface{}, error) {
	result := make([]interface{}, 0)
	err := json.Unmarshal([]byte(data), &result)
	return result, err
}
