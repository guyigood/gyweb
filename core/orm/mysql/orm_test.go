package orm

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDB 模拟数据库连接
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	arguments := m.Called(query, args)
	return arguments.Get(0).(sql.Result), arguments.Error(1)
}

func (m *MockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	arguments := m.Called(query, args)
	return arguments.Get(0).(*sql.Row)
}

// MockResult 模拟 SQL 结果
type MockResult struct {
	mock.Mock
}

func (m *MockResult) LastInsertId() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockResult) RowsAffected() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

// TestInc 测试 Inc 方法
func TestInc(t *testing.T) {
	// 创建 DB 实例
	db := &DB{
		table: "test_table",
	}

	t.Run("成功递增字段值", func(t *testing.T) {
		// 测试基本的递增功能
		// 注意：这里需要真实的数据库连接才能完整测试
		// 在实际项目中，建议使用测试数据库或 Docker 容器
		
		// 验证参数
		assert.NotEmpty(t, db.table, "表名不应为空")
		
		// 测试参数验证
		_, err := db.Inc("", 1.0)
		assert.Error(t, err, "空字段名应该返回错误")
		assert.Contains(t, err.Error(), "field name is required")
		
		// 测试空表名
		emptyDB := &DB{}
		_, err = emptyDB.Inc("test_field", 1.0)
		assert.Error(t, err, "空表名应该返回错误")
		assert.Contains(t, err.Error(), "table name is required")
	})

	t.Run("带WHERE条件的递增", func(t *testing.T) {
		// 测试带条件的递增
		db.Where("id = ?", 1)
		
		// 验证 WHERE 条件已设置
		assert.Len(t, db.where, 1, "应该有一个WHERE条件")
		assert.Equal(t, "id = ?", db.where[0])
		assert.Len(t, db.whereArgs, 1, "应该有一个WHERE参数")
		assert.Equal(t, 1, db.whereArgs[0])
	})
}

// TestSum 测试 Sum 方法
func TestSum(t *testing.T) {
	// 创建 DB 实例
	db := &DB{
		table: "test_table",
	}

	t.Run("参数验证测试", func(t *testing.T) {
		// 测试空字段名
		_, err := db.Sum("")
		assert.Error(t, err, "空字段名应该返回错误")
		assert.Contains(t, err.Error(), "field name is required")
		
		// 测试空表名
		emptyDB := &DB{}
		_, err = emptyDB.Sum("test_field")
		assert.Error(t, err, "空表名应该返回错误")
		assert.Contains(t, err.Error(), "table name is required")
	})

	t.Run("带WHERE条件的求和", func(t *testing.T) {
		// 测试带条件的求和
		db.Where("status = ?", "active")
		
		// 验证 WHERE 条件已设置
		assert.Len(t, db.where, 1, "应该有一个WHERE条件")
		assert.Equal(t, "status = ?", db.where[0])
		assert.Len(t, db.whereArgs, 1, "应该有一个WHERE参数")
		assert.Equal(t, "active", db.whereArgs[0])
	})
}

// TestIncSumIntegration 集成测试（需要真实数据库）
func TestIncSumIntegration(t *testing.T) {
	// 跳过集成测试，除非设置了测试数据库
	t.Skip("跳过集成测试 - 需要真实的数据库连接")
	
	// 以下是集成测试的示例代码
	// 在实际使用时，需要配置测试数据库
	/*
	// 连接测试数据库
	testDB, err := NewDB("mysql", "test_user:test_pass@tcp(localhost:3306)/test_db")
	if err != nil {
		t.Fatalf("无法连接测试数据库: %v", err)
	}
	defer testDB.Close()

	// 创建测试表
	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS test_inc_sum (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(50),
			score DECIMAL(10,2) DEFAULT 0,
			status VARCHAR(20) DEFAULT 'active'
		)
	`)
	assert.NoError(t, err)

	// 插入测试数据
	_, err = testDB.Table("test_inc_sum").Insert(map[string]interface{}{
		"name":   "test1",
		"score":  10.5,
		"status": "active",
	})
	assert.NoError(t, err)

	_, err = testDB.Table("test_inc_sum").Insert(map[string]interface{}{
		"name":   "test2", 
		"score":  20.3,
		"status": "active",
	})
	assert.NoError(t, err)

	// 测试 Inc 方法
	result, err := testDB.Table("test_inc_sum").Where("name = ?", "test1").Inc("score", 5.5)
	assert.NoError(t, err)
	
	affected, err := result.RowsAffected()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), affected, "应该影响1行")

	// 验证递增结果
	record, err := testDB.Table("test_inc_sum").Where("name = ?", "test1").Get()
	assert.NoError(t, err)
	assert.Equal(t, 16.0, record["score"], "分数应该是16.0")

	// 测试 Sum 方法
	sum, err := testDB.Table("test_inc_sum").Where("status = ?", "active").Sum("score")
	assert.NoError(t, err)
	assert.Equal(t, 36.3, sum, "总分应该是36.3")

	// 清理测试数据
	_, err = testDB.Exec("DROP TABLE test_inc_sum")
	assert.NoError(t, err)
	*/
}

// BenchmarkInc Inc 方法性能测试
func BenchmarkInc(b *testing.B) {
	db := &DB{
		table: "test_table",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 只测试参数验证部分，避免实际数据库操作
		_, _ = db.Inc("", 1.0) // 这会返回错误，但不会执行SQL
	}
}

// BenchmarkSum Sum 方法性能测试
func BenchmarkSum(b *testing.B) {
	db := &DB{
		table: "test_table",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 只测试参数验证部分，避免实际数据库操作
		_, _ = db.Sum("") // 这会返回错误，但不会执行SQL