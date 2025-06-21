package dbcommon

import (
	"fmt"
	"strings"
	"{project_name}/model"
	"{project_name}/public"

	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
	"github.com/guyigood/gyweb/core/utils/datatype"
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

	tableName := c.Param("table")
	if tableName == "" {
		c.Error(102, "表名不能为空")
		return
	}

	pageSizeStr := c.Query("page_size")
	pageSize := 20
	if pageSizeStr != "" {
		pageSize, _ = datatype.StrtoInt(pageSizeStr)
	}
	pageNoStr := c.Query("page_no")
	pageNo := 1
	if pageNoStr != "" {
		pageNo, _ = datatype.StrtoInt(pageNoStr)
	}

	tbinfo, err := public.GetTbInfoByTableName(tableName)
	if err != nil {
		c.Error(101, "参数错误")
		return
	}
	middleware.DebugVar("tbinfo", tbinfo)
	db := public.GetDb()
	countdb := public.GetDb()
	query := db.Table(tableName)
	countQuery := countdb.Table(tableName)
	if tbinfo.JoinTable != "" {
		query = query.Join(tbinfo.JoinTable).Select(tbinfo.JoinField + "," + tableName + ".*")
	}

	for _, v := range tbinfo.FdInfo {
		if v.FieldName == public.SysConfig.Database.LogicDeleteField {
			query.Where(public.SysConfig.Database.LogicDeleteField+"<>?", public.SysConfig.Database.LogicDeleteValue)
			countdb.Where(public.SysConfig.Database.LogicDeleteField+"<>?", public.SysConfig.Database.LogicDeleteValue)
		}
		if v.IsSearchable {
			searchkey := c.Query(v.FieldName)
			if searchkey == "" {
				continue
			}
			whereClause, whereArgs := buildConditions(v, searchkey)
			middleware.DebugVar("whereClause", whereClause)
			middleware.DebugVar("whereArgs", whereArgs)
			if whereClause != "" {
				query = query.Where(whereClause, whereArgs...)
				countQuery = countQuery.Where(whereClause, whereArgs...)
			}
		}
	}

	totalCount, err := countQuery.Count()
	if err != nil {
		c.Error(103, fmt.Sprintf("获取总数失败: %v", err))
		return
	}

	order := c.Query("order")
	sortBy := c.Query("sort_by")
	if sortBy != "" {
		query = query.OrderBy(fmt.Sprintf("%s %s", sortBy, order))

	} else {
		query = query.OrderBy(tbinfo.PrimaryKey + " DESC")
	}
	// 计算总页数
	totalPages := (totalCount + int64(pageSize) - 1) / int64(pageSize)

	// 添加分页
	offset := (pageNo - 1) * pageSize
	query = query.Limit(pageSize).Offset(offset)
	// 执行查询
	data, err := query.All()
	if err != nil {
		c.Error(104, fmt.Sprintf("查询失败: %v", err))
		return
	}

	// 返回分页结果，包含元数据信息
	result := gyarn.H{
		"data":        data,
		"total":       totalCount,
		"page":        pageNo,
		"page_size":   pageSize,
		"total_pages": totalPages,
		"has_next":    pageNo < int(totalPages),
		"has_prev":    pageNo > 1,
	}

	c.Success(result)
}

// buildSimpleConditions 构建简单的条件列表
func buildConditions(fdInfo model.GLobalFdInfo, value string) (string, []interface{}) {
	switch strings.ToLower(fdInfo.QueryType) {
	case "like":
		return fmt.Sprintf("%s LIKE ?", fdInfo.FieldName), []interface{}{"%" + value + "%"}
	case "eq":
		return fmt.Sprintf("%s = ?", fdInfo.FieldName), []interface{}{value}
	case "neq":
		return fmt.Sprintf("%s <> ?", fdInfo.FieldName), []interface{}{value}
	case "gt":
		return fmt.Sprintf("%s > ?", fdInfo.FieldName), []interface{}{value}
	case "gte":
		return fmt.Sprintf("%s >= ?", fdInfo.FieldName), []interface{}{value}
	case "lte":
		return fmt.Sprintf("%s <= ?", fdInfo.FieldName), []interface{}{value}
	case "lt":
		return fmt.Sprintf("%s < ?", fdInfo.FieldName), []interface{}{value}
	case "in":
		if fdInfo.FieldType == "varchar" {
			//将value逗号串变成带‘的逗号串
			value = strings.ReplaceAll(value, ",", "','")
			return fmt.Sprintf("%s IN (?)", fdInfo.FieldName), []interface{}{value}
		} else {
			return fmt.Sprintf("%s IN (?)", fdInfo.FieldName), []interface{}{value}
		}
	case "not_in":
		if fdInfo.FieldType == "varchar" {
			value = strings.ReplaceAll(value, ",", "','")
			return fmt.Sprintf("%s NOT IN (?)", fdInfo.FieldName), []interface{}{value}
		} else {
			return fmt.Sprintf("%s NOT IN (?)", fdInfo.FieldName), []interface{}{value}
		}
	case "between":
		values := strings.Split(value, ",")
		if len(values) != 2 {
			return "", nil
		}
		return fmt.Sprintf("%s BETWEEN ? AND ?", fdInfo.FieldName), []interface{}{values[0], values[1]}
	case "not_between":
		values := strings.Split(value, ",")
		if len(values) != 2 {
			return "", nil
		}
		return fmt.Sprintf("%s NOT BETWEEN ? AND ?", fdInfo.FieldName), []interface{}{values[0], values[1]}
	case "is_null":
		return fmt.Sprintf("%s IS NULL", fdInfo.FieldName), nil
	case "is_not_null":
		return fmt.Sprintf("%s IS NOT NULL", fdInfo.FieldName), nil
	case "is_empty":
		return fmt.Sprintf("%s = ''", fdInfo.FieldName), nil
	case "is_not_empty":
		return fmt.Sprintf("%s <> ''", fdInfo.FieldName), nil
	}
	return "", nil
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
	tableName := c.Param("table")
	if tableName == "" {
		c.Error(102, "表名不能为空")
		return
	}

	tbinfo, err := public.GetTbInfoByTableName(tableName)
	if err != nil {
		c.Error(101, "参数错误")
		return
	}
	db := public.GetDb()
	query := db.Table(tableName)

	if tbinfo.JoinTable != "" {
		query = query.Join(tbinfo.JoinTable).Select(tbinfo.JoinField + "," + tableName + ".*")
	}
	for _, v := range tbinfo.FdInfo {
		if v.FieldName == public.SysConfig.Database.LogicDeleteField {
			query.Where(public.SysConfig.Database.LogicDeleteField+"<>?", public.SysConfig.Database.LogicDeleteValue)
		}
		if v.IsSearchable {
			searchkey := c.Query(v.FieldName)
			if searchkey == "" {
				continue
			}
			whereClause, whereArgs := buildConditions(v, searchkey)
			if whereClause != "" {
				query = query.Where(whereClause, whereArgs...)
			}
		}
	}

	order := c.Query("order")
	sortBy := c.Query("sort_by")
	if sortBy != "" {
		query = query.OrderBy(fmt.Sprintf("%s %s", sortBy, order))

	} else {
		query = query.OrderBy(tbinfo.PrimaryKey + " DESC")
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
	webdata := make(map[string]interface{})
	if err := c.BindJSON(&webdata); err != nil {
		c.Error(101, "参数错误")
		return
	}

	tableName := c.Param("table")
	if tableName == "" {
		c.Error(102, "表名不能为空")
		return
	}
	tbinfo, err1 := public.GetTbInfoByTableName(tableName)
	if err1 != nil {
		c.Error(101, "参数错误")
		return
	}
	db := public.GetDb()
	for _, v := range tbinfo.FdInfo {
		if v.IsRequired && !v.IsPk {
			if _, ok := webdata[v.FieldName]; !ok {
				c.Error(104, fmt.Sprintf("缺少必填字段: %s", v.FieldName))
				return
			}
		}
		if v.IsUnique {
			if _, ok := webdata[v.FieldName]; ok {
				if datatype.TypetoStr(webdata[tbinfo.PrimaryKey]) == "" {
					count, err := db.Table(tableName).Where(v.FieldName+" = ?", webdata[v.FieldName]).Count()
					if err != nil {
						c.Error(104, fmt.Sprintf("查询失败: %v", err))
						return
					}
					if count > 0 {
						c.Error(104, fmt.Sprintf("字段 %s 不能重复", v.FieldName))
						return
					}
				} else {
					count, err := db.Table(tableName).Where(v.FieldName+" = ?", webdata[v.FieldName]).Where("id <> ?", webdata[tbinfo.PrimaryKey]).Count()
					if err != nil {
						c.Error(104, fmt.Sprintf("查询失败: %v", err))
						return
					}
					if count > 0 {
						c.Error(104, fmt.Sprintf("字段 %s 不能重复", v.FieldName))
						return
					}
				}
			}
		}
	}

	if datatype.TypetoStr(webdata[tbinfo.PrimaryKey]) == "" {
		_, err := db.Table(tableName).Insert(webdata)
		if err != nil {
			c.Error(104, fmt.Sprintf("插入失败: %v", err))
			return
		}
	} else {
		_, err := db.Table(tableName).Where(tbinfo.PrimaryKey+" = ?", webdata[tbinfo.PrimaryKey]).Update(webdata)
		if err != nil {
			c.Error(104, fmt.Sprintf("更新失败: %v", err))
			return
		}
	}
	c.Success(webdata)
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
	tableName := c.Param("table")
	if tableName == "" {
		c.Error(102, "表名不能为空")
		return
	}

	tbinfo, err1 := public.GetTbInfoByTableName(tableName)
	if err1 != nil {
		c.Error(101, "参数错误")
		return
	}

	id := c.Query("id")
	if id == "" {
		c.Error(102, "ID不能为空")
		return
	}
	is_del := true
	for _, v := range tbinfo.FdInfo {
		if v.FieldName == public.SysConfig.Database.LogicDeleteField {
			is_del = false
		}
	}
	db := public.GetDb()
	query := db.Table(tableName)
	query.Where(fmt.Sprintf("%s = ?", tbinfo.PrimaryKey), id)
	if is_del {
		_, err := query.Delete()
		if err != nil {
			c.Error(104, fmt.Sprintf("删除失败: %v", err))
			return
		}
	} else {
		query.Update(map[string]interface{}{public.SysConfig.Database.LogicDeleteField: public.SysConfig.Database.LogicDeleteValue})
	}
	c.Success("删除成功！")
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
	tableName := c.Param("table")
	if tableName == "" {
		c.Error(102, "表名不能为空")
		return
	}

	tbinfo, err1 := public.GetTbInfoByTableName(tableName)
	if err1 != nil {
		c.Error(101, "参数错误")
		return
	}

	id := c.Query("id")
	if id == "" {
		c.Error(102, "ID不能为空")
		return
	}

	db := public.GetDb()
	query := db.Table(tableName)
	if tbinfo.JoinTable != "" {
		query = query.Join(tbinfo.JoinTable).Select(tbinfo.JoinField + "," + tableName + ".*")
	}
	query.Where(fmt.Sprintf("%s = ?", tbinfo.PrimaryKey), id)
	data, err := query.Get()
	if err != nil {
		c.Error(104, fmt.Sprintf("查询失败: %v", err))
		return
	}

	c.Success(data)
}
