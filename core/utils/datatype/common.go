package datatype

import (
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
