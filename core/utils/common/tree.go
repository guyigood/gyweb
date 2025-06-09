package common

import (
	"fmt"
	"strconv"
)

// 构建树结构 - 统一处理int和string类型的ID
func BuildTree(data []map[string]interface{}, parentID interface{}, idKey, parentIDKey string) []map[string]interface{} {
	tree := make([]map[string]interface{}, 0)
	for _, v := range data {
		if compareValues(v[parentIDKey], parentID) {
			// 复制当前节点
			node := make(map[string]interface{})
			for k, val := range v {
				node[k] = val
			}

			// 递归构建子树
			children := BuildTree(data, v[idKey], idKey, parentIDKey)
			if len(children) > 0 {
				node["children"] = children
			}

			tree = append(tree, node)
		}
	}
	return tree
}

// 比较两个值是否相等（统一处理int、string等类型）
func compareValues(a, b interface{}) bool {
	if a == b {
		return true
	}

	// 尝试转换为相同类型进行比较
	aInt := convertToInt(a)
	bInt := convertToInt(b)
	if aInt != -1 && bInt != -1 {
		return aInt == bInt
	}

	// 如果不能转换为数字，则转换为字符串比较
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return aStr == bStr
}

// 统一转换为整数（支持int、float64、string类型）
func convertToInt(val interface{}) int {
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return -1 // 返回-1表示无法转换
}
