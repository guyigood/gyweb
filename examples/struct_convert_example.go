package main

import (
	"fmt"
	"log"
	"time"

	"../core/utils/common"
)

// 用户信息结构体
type User struct {
	ID       int       `json:"id" db:"user_id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Age      *int      `json:"age,omitempty"` // 指针类型
	CreatedAt time.Time `json:"created_at"`
	Profile  *Profile  `json:"profile,omitempty"` // 嵌套结构体指针
	Tags     []string  `json:"tags"`              // 切片
	Settings map[string]interface{} `json:"settings"` // map类型
	Ignored  string    `json:"-"`                    // 忽略字段
}

// 用户资料结构体
type Profile struct {
	Bio      string   `json:"bio"`
	Website  string   `json:"website"`
	Skills   []string `json:"skills"`
	Social   Social   `json:"social"` // 嵌套结构体
}

// 社交信息结构体
type Social struct {
	Github   string `json:"github"`
	Twitter  string `json:"twitter"`
	LinkedIn string `json:"linkedin"`
}

// 简单结构体（用于测试基本类型）
type SimpleStruct struct {
	IntField    int     `json:"int_field"`
	FloatField  float64 `json:"float_field"`
	StringField string  `json:"string_field"`
	BoolField   bool    `json:"bool_field"`
}

func main() {
	fmt.Println("=== gyweb 结构体转换工具演示 ===")
	fmt.Println()

	// 1. 测试 StructToMap - 简单结构体
	fmt.Println("1. 简单结构体转Map:")
	simple := SimpleStruct{
		IntField:    42,
		FloatField:  3.14,
		StringField: "Hello World",
		BoolField:   true,
	}

	simpleMap := common.StructToMap(simple)
	fmt.Printf("原始结构体: %+v\n", simple)
	fmt.Printf("转换后Map: %+v\n", simpleMap)
	fmt.Println()

	// 2. 测试 MapToStruct - 简单结构体
	fmt.Println("2. Map转简单结构体:")
	var newSimple SimpleStruct
	if err := common.MapToStruct(simpleMap, &newSimple); err != nil {
		log.Printf("转换失败: %v", err)
	} else {
		fmt.Printf("转换后结构体: %+v\n", newSimple)
	}
	fmt.Println()

	// 3. 测试复杂结构体转换
	fmt.Println("3. 复杂结构体转Map:")
	age := 25
	user := User{
		ID:    1,
		Name:  "张三",
		Email: "zhangsan@example.com",
		Age:   &age,
		CreatedAt: time.Now(),
		Profile: &Profile{
			Bio:     "全栈开发工程师",
			Website: "https://zhangsan.dev",
			Skills:  []string{"Go", "Java", "Python", "JavaScript"},
			Social: Social{
				Github:   "zhangsan",
				Twitter:  "@zhangsan",
				LinkedIn: "zhangsan-dev",
			},
		},
		Tags: []string{"developer", "golang", "backend"},
		Settings: map[string]interface{}{
			"theme":         "dark",
			"notifications": true,
			"language":      "zh-CN",
		},
		Ignored: "这个字段会被忽略",
	}

	userMap := common.StructToMap(user)
	fmt.Printf("用户结构体转Map成功，包含 %d 个字段\n", len(userMap))
	fmt.Printf("用户ID: %v\n", userMap["id"])
	fmt.Printf("用户名: %v\n", userMap["name"])
	fmt.Printf("年龄: %v\n", userMap["age"])
	fmt.Printf("标签: %v\n", userMap["tags"])
	fmt.Printf("设置: %v\n", userMap["settings"])
	if profile, ok := userMap["profile"].(map[string]interface{}); ok {
		fmt.Printf("个人资料技能: %v\n", profile["skills"])
		if social, ok := profile["social"].(map[string]interface{}); ok {
			fmt.Printf("GitHub: %v\n", social["github"])
		}
	}
	fmt.Println()

	// 4. 测试 Map 转复杂结构体
	fmt.Println("4. Map转复杂结构体:")
	var newUser User
	if err := common.MapToStruct(userMap, &newUser); err != nil {
		log.Printf("转换失败: %v", err)
	} else {
		fmt.Printf("转换成功!\n")
		fmt.Printf("用户ID: %d\n", newUser.ID)
		fmt.Printf("用户名: %s\n", newUser.Name)
		fmt.Printf("邮箱: %s\n", newUser.Email)
		if newUser.Age != nil {
			fmt.Printf("年龄: %d\n", *newUser.Age)
		}
		fmt.Printf("标签数量: %d\n", len(newUser.Tags))
		if newUser.Profile != nil {
			fmt.Printf("个人简介: %s\n", newUser.Profile.Bio)
			fmt.Printf("技能数量: %d\n", len(newUser.Profile.Skills))
			fmt.Printf("GitHub: %s\n", newUser.Profile.Social.Github)
		}
	}
	fmt.Println()

	// 5. 测试指针类型
	fmt.Println("5. 指针类型测试:")
	userPtr := &user
	ptrMap := common.StructToMap(userPtr)
	fmt.Printf("指针结构体转Map成功，字段数: %d\n", len(ptrMap))
	fmt.Println()

	// 6. 测试 nil 值处理
	fmt.Println("6. nil值处理测试:")
	nilMap := common.StructToMap(nil)
	fmt.Printf("nil结构体转Map结果: %v\n", nilMap)

	var nilUser *User
	nilPtrMap := common.StructToMap(nilUser)
	fmt.Printf("nil指针转Map结果: %v\n", nilPtrMap)
	fmt.Println()

	// 7. 测试错误处理
	fmt.Println("7. 错误处理测试:")
	
	// 测试非指针目标
	var testUser User
	err := common.MapToStruct(userMap, testUser) // 错误：应该传指针
	if err != nil {
		fmt.Printf("预期错误 - 非指针目标: %v\n", err)
	}

	// 测试 nil 指针
	var nilUserPtr *User
	err = common.MapToStruct(userMap, nilUserPtr) // 错误：nil指针
	if err != nil {
		fmt.Printf("预期错误 - nil指针: %v\n", err)
	}

	// 测试 nil map
	err = common.MapToStruct(nil, &newUser) // 错误：nil map
	if err != nil {
		fmt.Printf("预期错误 - nil map: %v\n", err)
	}
	fmt.Println()

	// 8. 测试部分字段转换
	fmt.Println("8. 部分字段转换测试:")
	partialMap := map[string]interface{}{
		"id":   999,
		"name": "李四",
		"email": "lisi@example.com",
		// 缺少其他字段
	}

	var partialUser User
	if err := common.MapToStruct(partialMap, &partialUser); err != nil {
		log.Printf("部分转换失败: %v", err)
	} else {
		fmt.Printf("部分转换成功: ID=%d, Name=%s, Email=%s\n", 
			partialUser.ID, partialUser.Name, partialUser.Email)
		fmt.Printf("未设置的字段保持零值: Age=%v, Tags=%v\n", 
			partialUser.Age, partialUser.Tags)
	}
	fmt.Println()

	// 9. 测试类型转换
	fmt.Println("9. 类型转换测试:")
	typeConvertMap := map[string]interface{}{
		"int_field":    "42",     // 字符串转int（需要手动处理）
		"float_field":  42,       // int转float64
		"string_field": 123,      // int转string（需要手动处理）
		"bool_field":   1,        // int转bool（需要手动处理）
	}

	var convertStruct SimpleStruct
	if err := common.MapToStruct(typeConvertMap, &convertStruct); err != nil {
		fmt.Printf("类型转换测试 - 部分转换失败（预期）: %v\n", err)
	} else {
		fmt.Printf("类型转换结果: %+v\n", convertStruct)
	}

	fmt.Println()
	fmt.Println("=== 演示完成 ===")
}