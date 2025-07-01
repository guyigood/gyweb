package main

import (
	"fmt"
	"log"
	"time"

	"github.com/guyigood/gyweb/core/orm/mysqlshard"
)

// User 用户模型
type User struct {
	ID        int64     `db:"id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// TableName 实现 Model 接口
func (u User) TableName() string {
	return "users"
}

func main() {
	// 数据库连接配置
	dsn := "root:password@tcp(localhost:3306)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	
	// 创建数据库连接
	db, err := mysqlshard.NewDB("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 设置分表阈值为10条记录（用于演示）
	db.SetMaxRecords(10)

	// 注册用户表结构
	userTableSQL := `
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
`
	db.RegisterTableStruct("users", userTableSQL)

	fmt.Println("=== 分表功能演示 ===")

	// 1. 批量插入数据，触发自动分表
	fmt.Println("\n1. 批量插入数据...")
	for i := 1; i <= 25; i++ {
		user := map[string]interface{}{
			"username": fmt.Sprintf("user_%d", i),
			"email":    fmt.Sprintf("user%d@example.com", i),
		}
		
		id, err := db.Table("users").Insert(user)
		if err != nil {
			log.Printf("Failed to insert user %d: %v", i, err)
			continue
		}
		
		if i%5 == 0 {
			fmt.Printf("Inserted %d users, last ID: %d\n", i, id)
		}
	}

	// 2. 查看分表信息
	fmt.Println("\n2. 查看分表信息...")
	info, err := db.GetShardInfo("users")
	if err != nil {
		log.Fatal("Failed to get shard info:", err)
	}

	fmt.Printf("基础表名: %s\n", info["base_table"])
	fmt.Printf("分表数量: %d\n", info["shard_count"])
	fmt.Printf("总记录数: %d\n", info["total_records"])
	fmt.Printf("每个分表最大记录数: %d\n", info["max_records_per_shard"])

	// 打印每个分表的记录数
	fmt.Println("\n各分表记录数:")
	tableCounts := info["table_counts"].(map[string]int64)
	for table, count := range tableCounts {
		fmt.Printf("  %s: %d 条记录\n", table, count)
	}

	// 3. 查询单个分表数据
	fmt.Println("\n3. 查询当前分表数据...")
	currentUsers, err := db.Table("users").Limit(5).All()
	if err != nil {
		log.Fatal("Failed to query current shard:", err)
	}

	fmt.Printf("当前分表前5条记录:\n")
	for _, user := range currentUsers {
		fmt.Printf("  ID: %v, Username: %v, Email: %v\n", 
			user["id"], user["username"], user["email"])
	}

	// 4. 跨分表查询
	fmt.Println("\n4. 跨分表查询...")
	allUsers, err := db.Table("users").AllFromAllShards()
	if err != nil {
		log.Fatal("Failed to query all shards:", err)
	}

	fmt.Printf("所有分表总记录数: %d\n", len(allUsers))

	// 5. 条件查询
	fmt.Println("\n5. 条件查询演示...")
	// 在当前分表中查询
	user, err := db.Table("users").Where("username = ?", "user_1").Get()
	if err != nil {
		fmt.Printf("在当前分表中未找到 user_1: %v\n", err)
	} else {
		fmt.Printf("在当前分表中找到 user_1: %v\n", user["email"])
	}

	// 6. 更新数据
	fmt.Println("\n6. 更新数据演示...")
	affected, err := db.Table("users").Where("username = ?", "user_1").Update(map[string]interface{}{
		"email": "updated_user1@example.com",
	})
	if err != nil {
		fmt.Printf("更新失败: %v\n", err)
	} else {
		fmt.Printf("更新了 %d 条记录\n", affected)
	}

	// 7. 统计所有分表记录数
	fmt.Println("\n7. 统计所有分表记录数...")
	totalCount, err := db.Table("users").CountFromAllShards()
	if err != nil {
		log.Fatal("Failed to count all shards:", err)
	}
	fmt.Printf("所有分表总记录数: %d\n", totalCount)

	// 8. 使用结构体模型
	fmt.Println("\n8. 使用结构体模型...")
	newUser := User{
		Username:  "struct_user",
		Email:     "struct@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := db.Table("users").Insert(newUser)
	if err != nil {
		log.Printf("Failed to insert struct user: %v", err)
	} else {
		fmt.Printf("使用结构体插入用户，ID: %d\n", id)
	}

	// 9. 事务演示
	fmt.Println("\n9. 事务演示...")
	err = db.Transaction(func(tx *mysqlshard.DB) error {
		// 在事务中插入多个用户
		for i := 100; i < 103; i++ {
			user := map[string]interface{}{
				"username": fmt.Sprintf("tx_user_%d", i),
				"email":    fmt.Sprintf("tx%d@example.com", i),
			}
			
			_, err := tx.Table("users").Insert(user)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("事务执行失败: %v\n", err)
	} else {
		fmt.Println("事务执行成功")
	}

	// 10. 最终统计
	fmt.Println("\n10. 最终统计...")
	finalInfo, err := db.GetShardInfo("users")
	if err != nil {
		log.Fatal("Failed to get final shard info:", err)
	}

	fmt.Printf("\n=== 最终分表统计 ===\n")
	fmt.Printf("分表数量: %d\n", finalInfo["shard_count"])
	fmt.Printf("总记录数: %d\n", finalInfo["total_records"])

	finalTableCounts := finalInfo["table_counts"].(map[string]int64)
	for table, count := range finalTableCounts {
		fmt.Printf("%s: %d 条记录\n", table, count)
	}

	fmt.Println("\n=== 演示完成 ===")
}

// 辅助函数：演示自定义分表配置
func demonstrateCustomConfig() {
	fmt.Println("\n=== 自定义分表配置演示 ===")
	
	// 创建自定义配置
	config := mysqlshard.NewShardConfig()
	config.MaxRecords = 5 // 每个分表最多5条记录
	
	// 预先注册表结构
	orderTableSQL := `
CREATE TABLE orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
`
	config.TableStructs["orders"] = orderTableSQL
	
	dsn := "root:password@tcp(localhost:3306)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := mysqlshard.NewDBWithShardConfig("mysql", dsn, config)
	if err != nil {
		log.Fatal("Failed to create DB with custom config:", err)
	}
	defer db.Close()
	
	// 插入订单数据
	for i := 1; i <= 12; i++ {
		order := map[string]interface{}{
			"user_id": i % 3 + 1,
			"amount":  float64(i) * 10.99,
			"status":  "pending",
		}
		
		_, err := db.Table("orders").Insert(order)
		if err != nil {
			log.Printf("Failed to insert order %d: %v", i, err)
		}
	}
	
	// 查看订单分表信息
	info, err := db.GetShardInfo("orders")
	if err != nil {
		log.Fatal("Failed to get orders shard info:", err)
	}
	
	fmt.Printf("订单表分表数量: %d\n", info["shard_count"])
	fmt.Printf("订单总数: %d\n", info["total_records"])
}