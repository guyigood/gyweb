package orm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	_ "github.com/go-sql-driver/mysql"
	"github.com/guyigood/gyweb/core/middleware"
)

// 修改说明：
// 为了统一 SQL 语句的记录和追踪机制，在所有执行 SQL 的方法中添加了相同的 SQL 记录逻辑。
// 这样可以确保通过 GetLastSql() 方法能够获取到最后执行的 SQL 语句，便于调试和日志追踪。
// 修改的方法包括：Insert、Update、Delete、Query、QueryRow、Exec 等。

// DB 数据库连接结构
type DB struct {
	db *sql.DB
	// 查询构建器字段
	table           string
	fields          []string
	where           []string
	whereArgs       []interface{}
	orderBy         string
	limit           int
	offset          int
	incVal          float64
	incField        string
	sumVal          float64
	sumField        string
	joins           []string
	groupBy         string
	LastSql, having string
}

// Model 基础模型接口
type Model interface {
	TableName() string
}

// MapModel 使用 map 实现的模型
type MapModel map[string]interface{}

// TableName 实现 Model 接口
func (m MapModel) TableName() string {
	if table, ok := m["_table"].(string); ok {
		return table
	}
	return ""
}

// NewDB 创建数据库连接
func NewDB(driver, dsn string) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	return &DB{db: db}, nil
}

// Table 指定表名
func (db *DB) Table(name string) *DB {
	db.resetQuery()
	db.table = name
	return db
}

func (db *DB) SetMaxIdleConns(n int) {
	db.db.SetMaxIdleConns(n)
}

func (db *DB) SetMaxOpenConns(n int) {
	db.db.SetMaxOpenConns(n)
}

func (db *DB) SetConnMaxLifetime(d time.Duration) {
	db.db.SetConnMaxLifetime(d)
}

func (db *DB) GetLastSql() string {
	return db.LastSql
}

// Inc 原子性递增字段值
// 实现 SQL: UPDATE table SET field=field+value WHERE conditions
// 参数:
//   field: 要递增的字段名
//   val: 递增的数值
// 返回: sql.Result 和 error，与标准 update 操作相同
func (db *DB) Inc(field string, val float64) (sql.Result, error) {
	// 参数验证
	if db.table == "" {
		return nil, fmt.Errorf("table name is required")
	}
	if field == "" {
		return nil, fmt.Errorf("field name is required")
	}

	// 构建 UPDATE SQL 语句
	sql := fmt.Sprintf("UPDATE %s SET %s = %s + ?", db.table, field, field)
	args := []interface{}{val}

	// 添加 WHERE 条件
	if len(db.where) > 0 {
		sql += " WHERE " + strings.Join(db.where, " AND ")
		args = append(args, db.whereArgs...)
	}

	// 调试输出
	middleware.DebugSQL(sql, args...)
	
	// 统一记录执行的 SQL 语句，便于调试和日志追踪
	if len(args) > 0 {
		db.LastSql = fmt.Sprintf(sql, args...)
	} else {
		db.LastSql = sql
	}

	// 执行 SQL 语句
	result, err := db.db.Exec(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to increment field %s: %w", field, err)
	}

	return result, nil
}

// Sum 计算指定字段的总和
// 实现 SQL: SELECT SUM(field) FROM table WHERE conditions
// 参数:
//   field: 要求和的字段名
// 返回: float64 总和值和 error，NULL 值会被处理为 0.0
func (db *DB) Sum(field string) (float64, error) {
	// 参数验证
	if db.table == "" {
		return 0.0, fmt.Errorf("table name is required")
	}
	if field == "" {
		return 0.0, fmt.Errorf("field name is required")
	}

	// 构建 SELECT SUM SQL 语句
	sql := fmt.Sprintf("SELECT COALESCE(SUM(%s), 0) FROM %s", field, db.table)
	var args []interface{}

	// 添加 WHERE 条件
	if len(db.where) > 0 {
		sql += " WHERE " + strings.Join(db.where, " AND ")
		args = append(args, db.whereArgs...)
	}

	// 调试输出
	middleware.DebugSQL(sql, args...)
	
	// 统一记录执行的 SQL 语句，便于调试和日志追踪
	if len(args) > 0 {
		db.LastSql = fmt.Sprintf(sql, args...)
	} else {
		db.LastSql = sql
	}

	// 执行查询
	var sum float64
	err := db.db.QueryRow(sql, args...).Scan(&sum)
	if err != nil {
		return 0.0, fmt.Errorf("failed to calculate sum for field %s: %w", field, err)
	}

	return sum, nil
}
// Model 使用模型
func (db *DB) Model(model interface{}) *DB {
	db.resetQuery()
	var tableName string
	if m, ok := model.(Model); ok {
		tableName = m.TableName()
	} else {
		if m, ok := model.(map[string]interface{}); ok {
			if table, ok := m["_table"].(string); ok {
				tableName = table
			}
		}
		if tableName == "" {
			t := reflect.TypeOf(model)
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			tableName = strings.ToLower(t.Name())
		}
	}
	db.table = tableName
	return db
}

// Select 指定查询字段
func (db *DB) Select(fields ...string) *DB {
	db.fields = fields
	return db
}

// Where 添加查询条件
func (db *DB) Where(condition string, args ...interface{}) *DB {
	db.where = append(db.where, condition)
	db.whereArgs = append(db.whereArgs, args...)
	return db
}

// OrderBy 指定排序
func (db *DB) OrderBy(order string) *DB {
	db.orderBy = order
	return db
}

// Limit 限制返回数量
func (db *DB) Limit(limit int) *DB {
	db.limit = limit
	return db
}

// Offset 指定偏移量
func (db *DB) Offset(offset int) *DB {
	db.offset = offset
	return db
}

// Join 添加连接查询
func (db *DB) Join(join string) *DB {
	db.joins = append(db.joins, join)
	return db
}

// GroupBy 指定分组
func (db *DB) GroupBy(group string) *DB {
	db.groupBy = group
	return db
}

// Having 添加分组条件
func (db *DB) Having(having string) *DB {
	db.having = having
	return db
}

// resetQuery 重置查询构建器
func (db *DB) resetQuery() {
	db.table = ""
	db.fields = nil
	db.where = nil
	db.whereArgs = nil
	db.orderBy = ""
	db.limit = 0
	db.offset = 0
	db.joins = nil
	db.groupBy = ""
	db.having = ""
}

// buildQuery 构建查询语句
func (db *DB) buildQuery() (string, []interface{}) {
	var sql strings.Builder
	var args []interface{}

	// SELECT 部分
	if len(db.fields) > 0 {
		sql.WriteString("SELECT " + strings.Join(db.fields, ", "))
	} else {
		sql.WriteString("SELECT *")
	}

	// FROM 部分
	sql.WriteString(" FROM " + db.table)

	// JOIN 部分
	if len(db.joins) > 0 {
		sql.WriteString(" " + strings.Join(db.joins, " "))
	}

	// WHERE 部分
	if len(db.where) > 0 {
		sql.WriteString(" WHERE " + strings.Join(db.where, " AND "))
		args = append(args, db.whereArgs...)
	}

	// GROUP BY 部分
	if db.groupBy != "" {
		sql.WriteString(" GROUP BY " + db.groupBy)
	}

	// HAVING 部分
	if db.having != "" {
		sql.WriteString(" HAVING " + db.having)
	}

	// ORDER BY 部分
	if db.orderBy != "" {
		sql.WriteString(" ORDER BY " + db.orderBy)
	}

	// LIMIT 和 OFFSET 部分
	if db.limit > 0 {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", db.limit))
		if db.offset > 0 {
			sql.WriteString(fmt.Sprintf(" OFFSET %d", db.offset))
		}
	}
	if len(args) > 0 {
		db.LastSql = fmt.Sprintf(sql.String(), args...)
	} else {
		db.LastSql = sql.String()
	}
	return sql.String(), args
}

// Get 获取单条记录到 map
func (db *DB) Get() (MapModel, error) {
	db.limit = 1
	sql, args := db.buildQuery()
	middleware.DebugSQL(sql, args...)
	rows, err := db.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if !rows.Next() {
		return nil, fmt.Errorf("no rows found")
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	result := make(MapModel)
	for i, col := range columns {
		val := values[i]
		if b, ok := val.([]byte); ok {
			result[strings.ToLower(col)] = string(b)
		} else {
			result[strings.ToLower(col)] = val
		}
	}

	return result, nil
}

// All 获取多条记录到 map 切片
func (db *DB) All() ([]MapModel, error) {
	sql, args := db.buildQuery()
	middleware.DebugSQL(sql, args...)
	rows, err := db.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []MapModel
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		result := make(MapModel)
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				result[strings.ToLower(col)] = string(b)
			} else {
				result[strings.ToLower(col)] = val
			}
		}
		results = append(results, result)
	}

	return results, nil
}

// 获取查询记录数
func (db *DB) Count() (int64, error) {
	db.fields = []string{"COUNT(*)"}
	sql, args := db.buildQuery()
	middleware.DebugSQL(sql, args...)
	rows, err := db.db.Query(sql, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, fmt.Errorf("no rows found")
	}

	var count int64
	if err := rows.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// Insert 插入数据（支持 map 和结构体）
func (db *DB) Insert(data interface{}) (sql.Result, error) {
	var fields []string
	var placeholders []string
	var args []interface{}

	switch v := data.(type) {
	case map[string]interface{}:
		for field, value := range v {
			if field != "_table" {
				fields = append(fields, field)
				placeholders = append(placeholders, "?")
				args = append(args, value)
			}
		}
	case Model:
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			column := field.Tag.Get("db")
			if column == "" {
				column = strings.ToLower(field.Name)
			}
			fields = append(fields, column)
			placeholders = append(placeholders, "?")
			args = append(args, val.Field(i).Interface())
		}
	default:
		return nil, fmt.Errorf("unsupported data type: %T", data)
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		db.table,
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "))

	middleware.DebugSQL(sql, args...)
	// 统一记录执行的 SQL 语句，便于调试和日志追踪
	if len(args) > 0 {
		db.LastSql = fmt.Sprintf(sql, args...)
	} else {
		db.LastSql = sql
	}
	result, err := db.db.Exec(sql, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Update 更新数据（支持 map 和结构体）
func (db *DB) Update(data interface{}) (sql.Result, error) {
	var sets []string
	var args []interface{}

	switch v := data.(type) {
	case map[string]interface{}:
		for field, value := range v {
			if field != "_table" {
				sets = append(sets, field+" = ?")
				args = append(args, value)
			}
		}
	case Model:
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			column := field.Tag.Get("db")
			if column == "" {
				column = strings.ToLower(field.Name)
			}
			sets = append(sets, column+" = ?")
			args = append(args, val.Field(i).Interface())
		}
	default:
		return nil, fmt.Errorf("unsupported data type: %T", data)
	}

	sql := fmt.Sprintf("UPDATE %s SET %s", db.table, strings.Join(sets, ", "))

	if len(db.where) > 0 {
		sql += " WHERE " + strings.Join(db.where, " AND ")
		args = append(args, db.whereArgs...)
	}

	middleware.DebugSQL(sql, args...)
	// 统一记录执行的 SQL 语句，便于调试和日志追踪
	if len(args) > 0 {
		db.LastSql = fmt.Sprintf(sql, args...)
	} else {
		db.LastSql = sql
	}
	result, err := db.db.Exec(sql, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Delete 删除数据
func (db *DB) Delete() (sql.Result, error) {
	sql := fmt.Sprintf("DELETE FROM %s", db.table)

	if len(db.where) > 0 {
		sql += " WHERE " + strings.Join(db.where, " AND ")
	}
 
	middleware.DebugSQL(sql, db.whereArgs...)
	// 统一记录执行的 SQL 语句，便于调试和日志追踪
	if len(db.whereArgs) > 0 {
		db.LastSql = fmt.Sprintf(sql, db.whereArgs...)
	} else {
		db.LastSql = sql
	}
	result, err := db.db.Exec(sql, db.whereArgs...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Transaction 执行事务函数
func (db *DB) Transaction(fn func(*DB) error) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}

	// 创建事务DB对象
	txDB := &DB{db: db.db}
	txDB.db = db.db

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(txDB); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	return db.db.Close()
}

// query 查询sql语句
func (db *DB) Query(sql string, args ...any) ([]MapModel, error) {
	middleware.DebugSQL(sql, args...)
	// 统一记录执行的 SQL 语句，便于调试和日志追踪
	if len(args) > 0 {
		db.LastSql = fmt.Sprintf(sql, args...)
	} else {
		db.LastSql = sql
	}
	rows, err := db.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []MapModel
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		result := make(MapModel)
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				result[strings.ToLower(col)] = string(b)
			} else {
				result[strings.ToLower(col)] = val
			}
		}
		results = append(results, result)
	}

	return results, nil
}

// exec 执行sql语句
func (db *DB) Exec(sql string, args ...any) (sql.Result, error) {
	middleware.DebugSQL(sql, args...)
	// 统一记录执行的 SQL 语句，便于调试和日志追踪
	if len(args) > 0 {
		db.LastSql = fmt.Sprintf(sql, args...)
	} else {
		db.LastSql = sql
	}
	return db.db.Exec(sql, args...)
}

// 雪花算法计算id
func (db *DB) SnowflakeID() (int64, error) {
	node, err := snowflake.NewNode(1)
	if err != nil {
		return 0, err
	}
	return node.Generate().Int64(), nil
}

func (db *DB) QueryRow(sql string, args ...any) (MapModel, error) {
	middleware.DebugSQL(sql, args...)
	// 统一记录执行的 SQL 语句，便于调试和日志追踪
	if len(args) > 0 {
		db.LastSql = fmt.Sprintf(sql, args...)
	} else {
		db.LastSql = sql
	}
	rows, err := db.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []MapModel
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		result := make(MapModel)
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				result[strings.ToLower(col)] = string(b)
			} else {
				result[strings.ToLower(col)] = val
			}
		}
		results = append(results, result)
		break
	}

	return results[0], nil
}

func (db *DB) Begin() (*sql.Tx, error) {
	tx, err := db.db.Begin()
	return tx, err
}

func (db *DB) Commit(tx *sql.Tx) error {
	return tx.Commit()
}

func (db *DB) Rollback(tx *sql.Tx) error {
	return tx.Rollback()
}
