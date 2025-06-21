package dbcommon

import (
	"fmt"
	"{project_name}/model"
	"{project_name}/public"
	"time"

	"github.com/guyigood/gyweb/core/gyarn"
	"github.com/guyigood/gyweb/core/middleware"
)

// @Summary 构建数据库表结构到sys表
// @Description 扫描当前数据库的所有表和字段信息，同步到sys_table_info和sys_field_info表中
// @Tags 数据库管理
// @Accept json
// @Produce json
// @Param database query string false "指定数据库名，默认使用当前连接的数据库"
// @Success 200 {object} map[string]interface{} "同步成功"
// @Failure 500 {object} map[string]interface{} "同步失败"
// @Router /api/db/build [get]
func BuildTable(c *gyarn.Context) {
	database := c.Query("database")
	db := public.GetDb()

	// 获取当前数据库名
	if database == "" {
		dbResult, err := db.Query("SELECT DATABASE() as db_name")
		if err != nil {
			c.Error(500, "获取当前数据库名失败: "+err.Error())
			return
		}
		if len(dbResult) > 0 {
			if dbName := dbResult[0]["db_name"]; dbName != nil {
				database = fmt.Sprintf("%v", dbName)
			}
		}
		if database == "" {
			c.Error(500, "无法获取当前数据库名")
			return
		}
	}

	// 检查并创建系统表
	err := ensureSystemTables(db)
	if err != nil {
		c.Error(500, "创建系统表失败: "+err.Error())
		return
	}

	// 获取表信息 - 添加调试信息
	fmt.Printf("[DEBUG] 查询数据库: %s\n", database)

	tables, err := getTableInfo(database)
	if err != nil {
		c.Error(500, "获取表信息失败: "+err.Error())
		return
	}

	fmt.Printf("[DEBUG] 找到 %d 个表\n", len(tables))

	if len(tables) == 0 {
		// 尝试查询所有数据库看看是否有权限问题
		allDbs, dbErr := public.GetDb().Query("SHOW DATABASES")
		if dbErr == nil {
			fmt.Printf("[DEBUG] 可访问的数据库: %v\n", allDbs)
		}

		// 尝试查询当前数据库中的所有表
		allTables, tableErr := public.GetDb().Query("SHOW TABLES")
		if tableErr == nil {
			fmt.Printf("[DEBUG] SHOW TABLES 结果: %v\n", allTables)
		}

		c.Error(500, fmt.Sprintf("数据库 '%s' 中没有找到任何表，请检查数据库名称和权限", database))
		return
	}

	syncStats := map[string]interface{}{
		"tables_processed": 0,
		"tables_added":     0,
		"tables_updated":   0,
		"fields_processed": 0,
		"fields_added":     0,
		"fields_updated":   0,
		"errors":           []string{},
	}

	// 同步表信息
	for _, table := range tables {
		// 跳过sys系列表本身
		if table.TableName == "sys_table_info" || table.TableName == "sys_field_info" || table.TableName == "sys_query_type" {
			continue
		}

		// 添加调试信息
		if table.TableName == "" {
			syncStats["errors"] = append(syncStats["errors"].([]string),
				"发现空表名，跳过处理")
			continue
		}
		middleware.DebugVar("table_name", table)
		syncStats["tables_processed"] = syncStats["tables_processed"].(int) + 1

		// 检查表是否已存在 - 改进错误处理
		existingTable, err := db.Table("sys_table_info").Where("table_name = ?", table.TableName).Get()

		// 区分"表不存在"和"真正的查询错误"
		if err != nil && err.Error() != "no rows found" {
			syncStats["errors"] = append(syncStats["errors"].([]string),
				fmt.Sprintf("查询表 %s 失败: %v", table.TableName, err))
			continue
		}

		// 如果是"no rows found"错误，说明表不存在，这是正常的
		if err != nil && err.Error() == "no rows found" {
			existingTable = nil
		}

		if existingTable == nil {
			// 新增表信息
			if table.TableComment == "" {
				table.TableComment = table.TableName
			}
			tableData := map[string]interface{}{
				"table_name":    table.TableName,
				"table_comment": table.TableComment,
				"module_name":   table.TableName,
				"status":        1,
				"create_time":   time.Now(),
				"update_time":   time.Now(),
			}

			_, err = db.Table("sys_table_info").Insert(tableData)
			if err != nil {
				syncStats["errors"] = append(syncStats["errors"].([]string),
					fmt.Sprintf("插入表 %s 失败: %v", table.TableName, err))
				continue
			}
			syncStats["tables_added"] = syncStats["tables_added"].(int) + 1
		} else {
			// 更新表信息
			updateData := map[string]interface{}{
				"table_comment": table.TableComment,
				"update_time":   time.Now(),
			}

			_, err = db.Table("sys_table_info").Where("table_name = ?", table.TableName).Update(updateData)
			if err != nil {
				syncStats["errors"] = append(syncStats["errors"].([]string),
					fmt.Sprintf("更新表 %s 失败: %v", table.TableName, err))
				continue
			}
			syncStats["tables_updated"] = syncStats["tables_updated"].(int) + 1
		}

		// 获取表的字段信息
		fields, err := getFieldInfo(database, table.TableName)
		if err != nil {
			syncStats["errors"] = append(syncStats["errors"].([]string),
				fmt.Sprintf("获取表 %s 字段信息失败: %v", table.TableName, err))
			continue
		}

		// 获取表ID
		tableInfo, err := db.Table("sys_table_info").Where("table_name = ?", table.TableName).Get()
		if err != nil || tableInfo == nil {
			syncStats["errors"] = append(syncStats["errors"].([]string),
				fmt.Sprintf("获取表 %s ID失败: %v", table.TableName, err))
			continue
		}

		tableID := fmt.Sprintf("%v", tableInfo["id"])

		// 同步字段信息
		for _, field := range fields {
			syncStats["fields_processed"] = syncStats["fields_processed"].(int) + 1

			// 检查字段是否已存在
			existingField, err := db.Table("sys_field_info").Where("table_id = ? AND field_name = ?",
				tableID, field.FieldName).Get()

			// 同样区分"记录不存在"和"真正的查询错误"
			if err != nil && err.Error() != "no rows found" {
				syncStats["errors"] = append(syncStats["errors"].([]string),
					fmt.Sprintf("查询字段 %s.%s 失败: %v", table.TableName, field.FieldName, err))
				continue
			}

			// 如果是"no rows found"错误，说明字段不存在，这是正常的
			if err != nil && err.Error() == "no rows found" {
				existingField = nil
			}

			// 转换字段类型和属性
			isPK := 0
			if field.ColumnKey == "PRI" {
				isPK = 1
			}

			isRequired := 0
			if field.IsNullable == "NO" {
				isRequired = 1
			}

			isUnique := 0
			if field.ColumnKey == "UNI" {
				isUnique = 1
			}

			// 默认显示设置
			showInList := 1
			showInAdd := 1
			showInEdit := 1
			if isPK == 1 || field.Extra == "auto_increment" {
				showInAdd = 0
				showInEdit = 0
			}

			// 根据字段类型设置查询类型
			queryType := "EQ"
			if field.FieldType == "varchar" || field.FieldType == "text" || field.FieldType == "longtext" {
				queryType = "LIKE"
			}

			// 表单类型
			formType := getFormType(field.FieldType)

			if existingField == nil {
				// 处理字段长度 - 防止超出 int 范围
				fieldLength := processFieldLength(field.FieldLength, field.FieldType)

				// 新增字段信息
				fieldData := map[string]interface{}{
					"table_id":         tableID,
					"field_name":       field.FieldName,
					"field_comment":    field.FieldComment,
					"field_type":       field.FieldType,
					"field_length":     fieldLength,
					"default_value":    field.DefaultValue,
					"is_pk":            isPK,
					"is_required":      isRequired,
					"is_unique":        isUnique,
					"show_in_list":     showInList,
					"show_in_add":      showInAdd,
					"show_in_edit":     showInEdit,
					"show_in_detail":   1,
					"is_searchable":    1,
					"query_type":       queryType,
					"list_sort_order":  field.OrdinalPos,
					"form_sort_order":  field.OrdinalPos,
					"query_sort_order": field.OrdinalPos,
					"form_type":        formType,
					"is_active":        1,
					"create_time":      time.Now(),
					"update_time":      time.Now(),
				}

				_, err = db.Table("sys_field_info").Insert(fieldData)
				if err != nil {
					syncStats["errors"] = append(syncStats["errors"].([]string),
						fmt.Sprintf("插入字段 %s.%s 失败: %v", table.TableName, field.FieldName, err))
					continue
				}
				syncStats["fields_added"] = syncStats["fields_added"].(int) + 1
			} else {
				// 处理字段长度 - 防止超出 int 范围
				/*fieldLength := processFieldLength(field.FieldLength, field.FieldType)

				// 更新字段信息
				updateData := map[string]interface{}{
					"field_comment": field.FieldComment,
					"field_type":    field.FieldType,
					"field_length":  fieldLength,
					"default_value": field.DefaultValue,
					"is_pk":         isPK,
					"is_required":   isRequired,
					"is_unique":     isUnique,
					"form_type":     formType,
					"update_time":   time.Now(),
				}

				fieldID := fmt.Sprintf("%v", existingField["id"])
				_, err = db.Table("sys_field_info").Where("id = ?", fieldID).Update(updateData)
				if err != nil {
					syncStats["errors"] = append(syncStats["errors"].([]string),
						fmt.Sprintf("更新字段 %s.%s 失败: %v", table.TableName, field.FieldName, err))
					continue
				}
				syncStats["fields_updated"] = syncStats["fields_updated"].(int) + 1*/
			}
		}
	}

	syncStats["success"] = true
	syncStats["message"] = "数据库表结构同步完成"
	syncStats["database"] = database
	syncStats["sync_time"] = time.Now().Format("2006-01-02 15:04:05")

	// 同步完成后刷新元数据缓存
	public.GetTbInfo()

	middleware.DebugVar("元数据缓存", "已成功刷新")
	syncStats["metadata_cache"] = "已成功刷新"

	c.JSON(200, syncStats)
}

// ensureSystemTables 确保系统表存在
func ensureSystemTables(db interface{}) error {
	// 这里可以检查系统表是否存在，如果不存在可以提示用户先运行SQL创建
	// 为了简化，这里暂时返回nil，假设系统表已经存在
	return nil
}

// getTableInfo 获取数据库表信息
func getTableInfo(database string) ([]model.TableInfo, error) {
	// 先尝试使用SHOW TABLES，这个更可靠
	showTablesQuery := "SHOW TABLES"
	fmt.Printf("[DEBUG] 执行 SHOW TABLES\n")

	rows, err := public.GetDb().Query(showTablesQuery)
	if err != nil {
		return nil, fmt.Errorf("SHOW TABLES 查询失败: %v", err)
	}

	fmt.Printf("[DEBUG] SHOW TABLES 返回 %d 行数据\n", len(rows))

	var tables []model.TableInfo
	for _, row := range rows {
		// SHOW TABLES 返回的字段名可能是 "Tables_in_数据库名"
		var tableName string
		for key, value := range row {
			if value != nil {
				tableName = fmt.Sprintf("%v", value)
				fmt.Printf("[DEBUG] 找到表: %s (字段名: %s)\n", tableName, key)
				break
			}
		}

		if tableName == "" || tableName == "<nil>" {
			continue
		}

		// 获取表注释 - 使用另一个查询
		tableComment := ""
		commentQuery := `
			SELECT table_comment 
			FROM information_schema.tables 
			WHERE table_schema = ? AND table_name = ?
		`
		commentRows, commentErr := public.GetDb().Query(commentQuery, database, tableName)
		if commentErr == nil && len(commentRows) > 0 {
			if commentRows[0]["table_comment"] != nil {
				tableComment = fmt.Sprintf("%v", commentRows[0]["table_comment"])
			}
		}

		table := model.TableInfo{
			TableName:    tableName,
			TableComment: tableComment,
			TableSchema:  database,
		}
		tables = append(tables, table)
	}

	fmt.Printf("[DEBUG] 最终找到 %d 个有效表\n", len(tables))
	return tables, nil
}

// getFieldInfo 获取表字段信息
func getFieldInfo(database, tableName string) ([]model.FieldInfo, error) {
	query := `
		SELECT 
			table_name,
			column_name as field_name,
			data_type as field_type,
			CASE 
				WHEN character_maximum_length IS NOT NULL THEN character_maximum_length
				WHEN numeric_precision IS NOT NULL THEN numeric_precision
				ELSE NULL
			END as field_length,
			column_default as default_value,
			COALESCE(column_comment, '') as field_comment,
			is_nullable,
			column_key,
			extra,
			ordinal_position
		FROM information_schema.columns 
		WHERE table_schema = ? AND table_name = ?
		ORDER BY ordinal_position
	`

	rows, err := public.GetDb().Query(query, database, tableName)
	if err != nil {
		return nil, fmt.Errorf("查询字段信息失败: %v", err)
	}

	var fields []model.FieldInfo
	for _, row := range rows {
		// 安全获取字符串字段
		fieldName := ""
		if row["field_name"] != nil {
			fieldName = fmt.Sprintf("%v", row["field_name"])
		}

		fieldType := ""
		if row["field_type"] != nil {
			fieldType = fmt.Sprintf("%v", row["field_type"])
		}

		fieldComment := ""
		if row["field_comment"] != nil {
			fieldComment = fmt.Sprintf("%v", row["field_comment"])
		}

		isNullable := ""
		if row["is_nullable"] != nil {
			isNullable = fmt.Sprintf("%v", row["is_nullable"])
		}

		columnKey := ""
		if row["column_key"] != nil {
			columnKey = fmt.Sprintf("%v", row["column_key"])
		}

		extra := ""
		if row["extra"] != nil {
			extra = fmt.Sprintf("%v", row["extra"])
		}

		// 跳过空字段名
		if fieldName == "" || fieldName == "<nil>" {
			continue
		}

		field := model.FieldInfo{
			TableName:    tableName,
			FieldName:    fieldName,
			FieldType:    fieldType,
			FieldLength:  row["field_length"],
			DefaultValue: row["default_value"],
			FieldComment: fieldComment,
			IsNullable:   isNullable,
			ColumnKey:    columnKey,
			Extra:        extra,
		}

		// 转换 ordinal_position 为 int
		if pos, ok := row["ordinal_position"].(int64); ok {
			field.OrdinalPos = int(pos)
		} else if pos, ok := row["ordinal_position"].(int); ok {
			field.OrdinalPos = pos
		} else {
			field.OrdinalPos = 1
		}

		fields = append(fields, field)
	}

	return fields, nil
}

// processFieldLength 处理字段长度，防止超出int范围
func processFieldLength(fieldLength interface{}, fieldType string) interface{} {
	if fieldLength == nil {
		return nil
	}

	// 对于text类型字段，通常长度很大，设置为固定值
	switch fieldType {
	case "text", "longtext", "mediumtext", "tinytext":
		return nil // text类型不需要长度限制
	case "json":
		return nil // json类型不需要长度限制
	}

	// 对于其他类型，检查长度是否超出int范围
	switch v := fieldLength.(type) {
	case int64:
		if v > 2147483647 { // int32最大值
			return 2147483647
		}
		if v < 0 {
			return nil
		}
		return int(v)
	case int:
		if v > 2147483647 {
			return 2147483647
		}
		if v < 0 {
			return nil
		}
		return v
	case float64:
		intVal := int64(v)
		if intVal > 2147483647 {
			return 2147483647
		}
		if intVal < 0 {
			return nil
		}
		return int(intVal)
	default:
		return fieldLength
	}
}

// getFormType 根据数据库字段类型返回表单控件类型
func getFormType(fieldType string) string {
	switch fieldType {
	case "varchar", "char":
		return "input"
	case "text", "longtext", "mediumtext":
		return "textarea"
	case "int", "bigint", "smallint", "tinyint":
		return "number"
	case "decimal", "float", "double":
		return "number"
	case "date":
		return "date"
	case "datetime", "timestamp":
		return "datetime"
	case "time":
		return "time"
	case "enum":
		return "select"
	case "json":
		return "textarea"
	default:
		return "input"
	}
}
