package common

// 构建树结构
func BuildTree(data []map[string]interface{}, parentID, idKey, parentIDKey string) []map[string]interface{} {
	tree := make([]map[string]interface{}, 0)
	for _, v := range data {
		if v[parentIDKey] == parentID {
			tree = append(tree, v)
			tree = append(tree, BuildTree(data, v[idKey].(string), idKey, parentIDKey)...)
		}
	}
	return tree
}
