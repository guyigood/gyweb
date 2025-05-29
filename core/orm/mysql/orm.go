package orm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DB 数据库连接结构
type DB struct {
	db *sql.DB
}

// Tx 事务结构
type Tx struct {
	tx *sql.Tx
	db *DB
}

// Query 查询构建器
type Query struct {
	db        *DB
	tx        *Tx
	table     string
	fields    []string
	where     []string
	whereArgs []interface{}
	orderBy   string
	limit     int
	offset    int
	joins     []string
	groupBy   string
	having    string
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

// Begin 开始事务
func (db *DB) Begin() (*Tx, error) {
	tx, err := db.db.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, db: db}, nil
}

// Commit 提交事务
func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

// Rollback 回滚事务
func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}

// Table 指定表名（事务版本）
func (tx *Tx) Table(name string) *Query {
	return &Query{
		tx:    tx,
		table: name,
	}
}

// Model 使用模型（事务版本）
func (tx *Tx) Model(model interface{}) *Query {
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

	return &Query{
		tx:    tx,
		table: tableName,
	}
}

// Select 指定查询字段
func (q *Query) Select(fields ...string) *Query {
	q.fields = fields
	return q
}

// Where 添加查询条件
func (q *Query) Where(condition string, args ...interface{}) *Query {
	q.where = append(q.where, condition)
	q.whereArgs = append(q.whereArgs, args...)
	return q
}

// OrderBy 指定排序
func (q *Query) OrderBy(order string) *Query {
	q.orderBy = order
	return q
}

// Limit 限制返回数量
func (q *Query) Limit(limit int) *Query {
	q.limit = limit
	return q
}

// Offset 指定偏移量
func (q *Query) Offset(offset int) *Query {
	q.offset = offset
	return q
}

// Join 添加连接查询
func (q *Query) Join(join string) *Query {
	q.joins = append(q.joins, join)
	return q
}

// GroupBy 指定分组
func (q *Query) GroupBy(group string) *Query {
	q.groupBy = group
	return q
}

// Having 添加分组条件
func (q *Query) Having(having string) *Query {
	q.having = having
	return q
}

// buildQuery 构建查询语句
func (q *Query) buildQuery() (string, []interface{}) {
	var sql strings.Builder
	var args []interface{}

	// SELECT 部分
	if len(q.fields) > 0 {
		sql.WriteString("SELECT " + strings.Join(q.fields, ", "))
	} else {
		sql.WriteString("SELECT *")
	}

	// FROM 部分
	sql.WriteString(" FROM " + q.table)

	// JOIN 部分
	if len(q.joins) > 0 {
		sql.WriteString(" " + strings.Join(q.joins, " "))
	}

	// WHERE 部分
	if len(q.where) > 0 {
		sql.WriteString(" WHERE " + strings.Join(q.where, " AND "))
		args = append(args, q.whereArgs...)
	}

	// GROUP BY 部分
	if q.groupBy != "" {
		sql.WriteString(" GROUP BY " + q.groupBy)
	}

	// HAVING 部分
	if q.having != "" {
		sql.WriteString(" HAVING " + q.having)
	}

	// ORDER BY 部分
	if q.orderBy != "" {
		sql.WriteString(" ORDER BY " + q.orderBy)
	}

	// LIMIT 和 OFFSET 部分
	if q.limit > 0 {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", q.limit))
		if q.offset > 0 {
			sql.WriteString(fmt.Sprintf(" OFFSET %d", q.offset))
		}
	}

	return sql.String(), args
}

// 修改所有数据库操作方法以支持事务
func (q *Query) execQuery(sql string, args ...interface{}) (*sql.Rows, error) {
	if q.tx != nil {
		return q.tx.tx.Query(sql, args...)
	}
	return q.db.db.Query(sql, args...)
}

func (q *Query) execRow(sql string, args ...interface{}) *sql.Row {
	if q.tx != nil {
		return q.tx.tx.QueryRow(sql, args...)
	}
	return q.db.db.QueryRow(sql, args...)
}

func (q *Query) exec(sql string, args ...interface{}) (sql.Result, error) {
	if q.tx != nil {
		return q.tx.tx.Exec(sql, args...)
	}
	return q.db.db.Exec(sql, args...)
}

// Get 获取单条记录到 map
func (q *Query) Get() (MapModel, error) {
	q.limit = 1
	sql, args := q.buildQuery()
	rows, err := q.execQuery(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// 准备接收数据的切片
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

	// 将结果转换为 map
	result := make(MapModel)
	for i, col := range columns {
		val := values[i]
		if b, ok := val.([]byte); ok {
			result[col] = string(b)
		} else {
			result[col] = val
		}
	}

	return result, nil
}

// All 获取多条记录到 map 切片
func (q *Query) All() ([]MapModel, error) {
	sql, args := q.buildQuery()
	rows, err := q.execQuery(sql, args...)
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
		// 准备接收数据的切片
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// 将结果转换为 map
		result := make(MapModel)
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				result[col] = string(b)
			} else {
				result[col] = val
			}
		}
		results = append(results, result)
	}

	return results, nil
}

// Insert 插入数据（支持 map 和结构体）
func (q *Query) Insert(data interface{}) (int64, error) {
	var fields []string
	var placeholders []string
	var args []interface{}

	switch v := data.(type) {
	case map[string]interface{}:
		for field, value := range v {
			if field != "_table" { // 忽略表名字段
				fields = append(fields, field)
				placeholders = append(placeholders, "?")
				args = append(args, value)
			}
		}
	case Model:
		// 使用反射获取结构体字段
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			// 获取字段标签或字段名
			column := field.Tag.Get("db")
			if column == "" {
				column = strings.ToLower(field.Name)
			}
			fields = append(fields, column)
			placeholders = append(placeholders, "?")
			args = append(args, val.Field(i).Interface())
		}
	default:
		return 0, fmt.Errorf("unsupported data type: %T", data)
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		q.table,
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "))

	result, err := q.exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// Update 更新数据（支持 map 和结构体）
func (q *Query) Update(data interface{}) (int64, error) {
	var sets []string
	var args []interface{}

	switch v := data.(type) {
	case map[string]interface{}:
		for field, value := range v {
			if field != "_table" { // 忽略表名字段
				sets = append(sets, field+" = ?")
				args = append(args, value)
			}
		}
	case Model:
		// 使用反射获取结构体字段
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			// 获取字段标签或字段名
			column := field.Tag.Get("db")
			if column == "" {
				column = strings.ToLower(field.Name)
			}
			sets = append(sets, column+" = ?")
			args = append(args, val.Field(i).Interface())
		}
	default:
		return 0, fmt.Errorf("unsupported data type: %T", data)
	}

	sql := fmt.Sprintf("UPDATE %s SET %s", q.table, strings.Join(sets, ", "))

	if len(q.where) > 0 {
		sql += " WHERE " + strings.Join(q.where, " AND ")
		args = append(args, q.whereArgs...)
	}

	result, err := q.exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Delete 删除数据
func (q *Query) Delete() (int64, error) {
	sql := fmt.Sprintf("DELETE FROM %s", q.table)

	if len(q.where) > 0 {
		sql += " WHERE " + strings.Join(q.where, " AND ")
	}

	result, err := q.exec(sql, q.whereArgs...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Transaction 执行事务函数
func (db *DB) Transaction(fn func(*Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	return db.db.Close()
}
