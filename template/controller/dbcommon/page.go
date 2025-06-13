package dbcommon

import (
	"fmt"
	"strings"
	"{project_name}/model"
	"{project_name}/public"

	"github.com/guyigood/gyweb/core/gyarn"
)

// Page 分页查询数据
// @Summary 分页查询数据
// @Tags 数据库通用操作
// @Accept json
// @Produce json
// @RequestBody {object} model.Page
// @Success 200 {object} PageResponse "查询结果"
// @Failure 101 {object} ErrorResponse "参数错误"
// @Failure 102 {object} ErrorResponse "表名不能为空"
// @Failure 103 {object} ErrorResponse "获取总数失败"
// @Failure 104 {object} ErrorResponse "查询失败"
// @Router /api/db/page [post]
func Page(c *gyarn.Context) {
	var req model.Page

	if err := c.BindJSON(&req); err != nil {
		c.Error(101, "参数错误")
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.TableName == "" {
		c.Error(102, "表名不能为空")
		return
	}

	db := public.Db
	query := db.Table(req.TableName)

	// 构建查询条件
	whereClause, whereArgs := buildWhereConditions(req)

	// 应用 WHERE 条件
	if whereClause != "" {
		query = query.Where(whereClause, whereArgs...)
	}

	// 获取总记录数（用于计算总页数）
	countQuery := db.Table(req.TableName)
	if whereClause != "" {
		countQuery = countQuery.Where(whereClause, whereArgs...)
	}

	totalCount, err := countQuery.Count()
	if err != nil {
		c.Error(103, fmt.Sprintf("获取总数失败: %v", err))
		return
	}

	// 计算总页数
	totalPages := (totalCount + int64(req.PageSize) - 1) / int64(req.PageSize)

	// 添加排序
	if req.SortBy != "" {
		order := "ASC"
		if strings.ToUpper(req.Order) == "DESC" {
			order = "DESC"
		}
		query = query.OrderBy(fmt.Sprintf("%s %s", req.SortBy, order))
	}

	// 添加分页
	offset := (req.Page - 1) * req.PageSize
	query = query.Limit(req.PageSize).Offset(offset)

	// 执行查询
	data, err := query.All()
	if err != nil {
		c.Error(104, fmt.Sprintf("查询失败: %v", err))
		return
	}

	// 返回分页结果
	result := gyarn.H{
		"data":        data,
		"total":       totalCount,
		"page":        req.Page,
		"page_size":   req.PageSize,
		"total_pages": totalPages,
		"has_next":    req.Page < int(totalPages),
		"has_prev":    req.Page > 1,
	}

	c.Success(result)
}

// buildWhereConditions 构建复杂的 WHERE 条件，支持 AND 和 OR 逻辑
func buildWhereConditions(req model.Page) (string, []interface{}) {
	var allConditions []string
	var allArgs []interface{}

	// 处理简单的 Filters（向后兼容）
	if len(req.Filters) > 0 {
		simpleConditions, simpleArgs := buildSimpleConditions(req.Filters)
		if len(simpleConditions) > 0 {
			// 简单条件用 AND 连接
			simpleWhere := strings.Join(simpleConditions, " AND ")
			allConditions = append(allConditions, "("+simpleWhere+")")
			allArgs = append(allArgs, simpleArgs...)
		}
	}

	// 处理复杂的 FilterGroups
	for _, group := range req.FilterGroups {
		groupConditions, groupArgs := buildSimpleConditions(group.Filters)
		if len(groupConditions) > 0 {
			logic := strings.ToUpper(group.Logic)
			if logic != "OR" && logic != "AND" {
				logic = "AND" // 默认使用 AND
			}

			// 组内条件用指定的逻辑连接
			groupWhere := strings.Join(groupConditions, " "+logic+" ")
			allConditions = append(allConditions, "("+groupWhere+")")
			allArgs = append(allArgs, groupArgs...)
		}
	}

	// 所有条件组之间用 AND 连接
	if len(allConditions) > 0 {
		finalWhere := strings.Join(allConditions, " AND ")
		return finalWhere, allArgs
	}

	return "", nil
}

// buildSimpleConditions 构建简单的条件列表
func buildSimpleConditions(filters map[string]model.PageFilter) ([]string, []interface{}) {
	var conditions []string
	var args []interface{}

	for field, filter := range filters {
		if filter.Value == nil {
			continue
		}

		switch strings.ToLower(filter.Operator) {
		case "like":
			conditions = append(conditions, fmt.Sprintf("%s LIKE ?", field))
			args = append(args, fmt.Sprintf("%%%v%%", filter.Value))

		case "=", "eq":
			conditions = append(conditions, fmt.Sprintf("%s = ?", field))
			args = append(args, filter.Value)

		case ">", "gt":
			conditions = append(conditions, fmt.Sprintf("%s > ?", field))
			args = append(args, filter.Value)

		case ">=", "gte":
			conditions = append(conditions, fmt.Sprintf("%s >= ?", field))
			args = append(args, filter.Value)

		case "<", "lt":
			conditions = append(conditions, fmt.Sprintf("%s < ?", field))
			args = append(args, filter.Value)

		case "<=", "lte":
			conditions = append(conditions, fmt.Sprintf("%s <= ?", field))
			args = append(args, filter.Value)

		case "!=", "<>", "ne":
			conditions = append(conditions, fmt.Sprintf("%s != ?", field))
			args = append(args, filter.Value)

		case "in":
			// 处理 IN 查询，支持数组或逗号分隔的字符串
			switch v := filter.Value.(type) {
			case []interface{}:
				if len(v) > 0 {
					placeholders := strings.Repeat("?,", len(v))
					placeholders = placeholders[:len(placeholders)-1] // 移除最后一个逗号
					conditions = append(conditions, fmt.Sprintf("%s IN (%s)", field, placeholders))
					args = append(args, v...)
				}
			case string:
				// 如果是字符串，按逗号分割
				values := strings.Split(v, ",")
				if len(values) > 0 {
					placeholders := strings.Repeat("?,", len(values))
					placeholders = placeholders[:len(placeholders)-1]
					conditions = append(conditions, fmt.Sprintf("%s IN (%s)", field, placeholders))
					for _, val := range values {
						args = append(args, strings.TrimSpace(val))
					}
				}
			}

		case "not in", "notin":
			// 处理 NOT IN 查询
			switch v := filter.Value.(type) {
			case []interface{}:
				if len(v) > 0 {
					placeholders := strings.Repeat("?,", len(v))
					placeholders = placeholders[:len(placeholders)-1]
					conditions = append(conditions, fmt.Sprintf("%s NOT IN (%s)", field, placeholders))
					args = append(args, v...)
				}
			case string:
				values := strings.Split(v, ",")
				if len(values) > 0 {
					placeholders := strings.Repeat("?,", len(values))
					placeholders = placeholders[:len(placeholders)-1]
					conditions = append(conditions, fmt.Sprintf("%s NOT IN (%s)", field, placeholders))
					for _, val := range values {
						args = append(args, strings.TrimSpace(val))
					}
				}
			}

		case "between":
			// 处理 BETWEEN 查询，期望值是包含两个元素的数组
			if values, ok := filter.Value.([]interface{}); ok && len(values) == 2 {
				conditions = append(conditions, fmt.Sprintf("%s BETWEEN ? AND ?", field))
				args = append(args, values[0], values[1])
			}

		case "is null", "isnull":
			conditions = append(conditions, fmt.Sprintf("%s IS NULL", field))

		case "is not null", "isnotnull":
			conditions = append(conditions, fmt.Sprintf("%s IS NOT NULL", field))

		case "startswith", "starts_with":
			conditions = append(conditions, fmt.Sprintf("%s LIKE ?", field))
			args = append(args, fmt.Sprintf("%v%%", filter.Value))

		case "endswith", "ends_with":
			conditions = append(conditions, fmt.Sprintf("%s LIKE ?", field))
			args = append(args, fmt.Sprintf("%%%v", filter.Value))

		default:
			// 对于不支持的操作符，记录日志或忽略
			fmt.Printf("Unsupported operator: %s for field: %s\n", filter.Operator, field)
		}
	}

	return conditions, args
}

// List 列表查询数据
// @Summary 列表查询数据
// @Description 支持复杂条件查询的通用列表接口，不分页返回所有符合条件的数据
// @Tags 数据库通用操作
// @Accept json
// @Produce json
// @RequestBody {object} model.Page
// @Success 200 {object} ListResponse "查询结果"
// @Failure 101 {object} ErrorResponse "参数错误"
// @Failure 102 {object} ErrorResponse "表名不能为空"
// @Failure 104 {object} ErrorResponse "查询失败"
// @Router /api/db/list [post]
func List(c *gyarn.Context) {
	var req model.Page

	if err := c.BindJSON(&req); err != nil {
		c.Error(101, "参数错误")
		return
	}

	if req.TableName == "" {
		c.Error(102, "表名不能为空")
		return
	}

	db := public.Db
	query := db.Table(req.TableName)

	// 构建查询条件
	whereClause, whereArgs := buildWhereConditions(req)

	// 应用 WHERE 条件
	if whereClause != "" {
		query = query.Where(whereClause, whereArgs...)
	}

	// 添加排序
	if req.SortBy != "" {
		order := "ASC"
		if strings.ToUpper(req.Order) == "DESC" {
			order = "DESC"
		}
		query = query.OrderBy(fmt.Sprintf("%s %s", req.SortBy, order))
	}

	// 执行查询
	data, err := query.All()
	if err != nil {
		c.Error(104, fmt.Sprintf("查询失败: %v", err))
		return
	}
	c.Success(data)
}

// Save 保存数据
// @Summary 保存数据
// @Description 通用数据保存接口，支持新增和更新操作，自动根据是否有id字段判断操作类型
// @Tags 数据库通用操作
// @Accept json
// @Produce json
// @RequestBody {object} model.SaveData
// @Success 200 {object} SaveResponse "保存成功，返回保存后的数据"
// @Failure 101 {object} ErrorResponse "参数错误"
// @Failure 102 {object} ErrorResponse "表名不能为空"
// @Failure 103 {object} ErrorResponse "保存数据不能为空"
// @Failure 104 {object} ErrorResponse "缺少必填字段"
// @Failure 105 {object} ErrorResponse "必填字段不能为空"
// @Failure 106 {object} ErrorResponse "没有需要更新的数据"
// @Failure 107 {object} ErrorResponse "更新失败"
// @Failure 108 {object} ErrorResponse "获取更新后数据失败"
// @Failure 109 {object} ErrorResponse "插入失败"
// @Router /api/db/save [post]
func Save(c *gyarn.Context) {
	var req model.SaveData

	if err := c.BindJSON(&req); err != nil {
		c.Error(101, "参数错误")
		return
	}

	if req.TableName == "" {
		c.Error(102, "表名不能为空")
		return
	}

	if req.Data == nil || len(req.Data) == 0 {
		c.Error(103, "保存数据不能为空")
		return
	}

	// 验证必填字段
	for _, field := range req.Required {
		value, exists := req.Data[field]
		if !exists {
			c.Error(104, fmt.Sprintf("缺少必填字段: %s", field))
			return
		}

		// 检查字段值是否为空
		if value == nil || value == "" {
			c.Error(105, fmt.Sprintf("必填字段不能为空: %s", field))
			return
		}
	}

	// 过滤数据：只保留 Required 和 Optional 字段
	filteredData := make(map[string]interface{})
	allowedFields := make(map[string]bool)

	// 添加必填字段到允许列表
	for _, field := range req.Required {
		allowedFields[field] = true
	}

	// 添加可选字段到允许列表
	for _, field := range req.Optional {
		allowedFields[field] = true
	}

	// 过滤数据
	for key, value := range req.Data {
		if allowedFields[key] {
			filteredData[key] = value
		}
	}

	db := public.Db

	// 判断是插入还是更新操作
	idValue, hasId := filteredData["id"]
	isUpdate := hasId && idValue != nil && idValue != "" && idValue != 0

	var err error
	var result interface{}

	if isUpdate {
		// 更新操作
		id := idValue
		delete(filteredData, "id") // 移除 id 字段，避免更新 id

		if len(filteredData) == 0 {
			c.Error(106, "没有需要更新的数据")
			return
		}

		// 更新操作
		query := db.Table(req.TableName)

		result, err = query.Where("id = ?", id).Update(filteredData)
		if err != nil {
			c.Error(107, fmt.Sprintf("更新失败: %v", err))
			return
		}

		// 返回更新后的数据
		updatedData, getErr := db.Table(req.TableName).Where("id = ?", id).Get()
		if getErr != nil {
			c.Error(108, fmt.Sprintf("获取更新后数据失败: %v", getErr))
			return
		}

		c.Success(updatedData)
	} else {
		// 插入操作
		result, err = db.Table(req.TableName).Insert(filteredData)
		if err != nil {
			c.Error(109, fmt.Sprintf("插入失败: %v", err))
			return
		}

		// 获取插入的 ID（如果有的话）
		var insertId interface{}
		if insertResult, ok := result.(map[string]interface{}); ok {
			if id, exists := insertResult["last_insert_id"]; exists {
				insertId = id
			}
		}

		// 返回插入后的数据
		var insertedData interface{}
		if insertId != nil && insertId != 0 {
			insertedData, err = db.Table(req.TableName).Where("id = ?", insertId).Get()
			if err != nil {
				// 如果获取失败，至少返回插入的数据
				insertedData = filteredData
			}
		} else {
			insertedData = filteredData
		}

		c.Success(insertedData)
	}
}

// Delete 删除数据
// @Summary 删除数据
// @Description 根据ID删除指定表中的数据记录
// @Tags 数据库通用操作
// @Accept json
// @Produce json
// @Param table query string true "表名" example("users")
// @Param id query string true "要删除的记录ID" example("1")
// @Success 200 {object} DeleteResponse "删除成功"
// @Failure 104 {object} ErrorResponse "删除失败"
// @Router /api/db/delete [post]
func Delete(c *gyarn.Context) {
	db := public.Db
	query := db.Table(c.Query("table"))
	query.Where("id = ?", c.Query("id"))
	_, err := query.Delete()
	if err != nil {
		c.Error(104, fmt.Sprintf("删除失败: %v", err))
		return
	}
	c.Success(nil)
}

// Detail 获取详情数据
// @Summary 获取详情数据
// @Description 根据ID获取指定表中单条记录的详细信息
// @Tags 数据库通用操作
// @Accept json
// @Produce json
// @Param table query string true "表名" example("users")
// @Param id query string true "记录ID" example("1")
// @Success 200 {object} DetailResponse "查询成功，返回记录详情"
// @Failure 104 {object} ErrorResponse "查询失败"
// @Router /api/db/detail [get]
func Detail(c *gyarn.Context) {
	db := public.Db
	query := db.Table(c.Query("table"))
	query.Where("id = ?", c.Query("id"))
	data, err := query.Get()
	if err != nil {
		c.Error(104, fmt.Sprintf("查询失败: %v", err))
		return
	}
	c.Success(data)
}
