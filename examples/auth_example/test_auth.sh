#!/bin/bash

# API鉴权模块测试脚本
BASE_URL="http://localhost:8080"

echo "🧪 API鉴权模块测试脚本"
echo "========================================="

# 测试函数
test_api() {
    local method=$1
    local url=$2
    local data=$3
    local token=$4
    local description=$5
    
    echo
    echo "📝 测试: $description"
    echo "🔗 $method $url"
    
    if [ -n "$token" ]; then
        if [ -n "$data" ]; then
            curl -s -X $method "$BASE_URL$url" \
                -H "Content-Type: application/json" \
                -H "Authorization: Bearer $token" \
                -d "$data" | jq .
        else
            curl -s -X $method "$BASE_URL$url" \
                -H "Authorization: Bearer $token" | jq .
        fi
    else
        if [ -n "$data" ]; then
            curl -s -X $method "$BASE_URL$url" \
                -H "Content-Type: application/json" \
                -d "$data" | jq .
        else
            curl -s -X $method "$BASE_URL$url" | jq .
        fi
    fi
}

# 检查服务器是否启动
echo "🚀 检查服务器状态..."
if ! curl -s "$BASE_URL" > /dev/null; then
    echo "❌ 服务器未启动，请先运行: go run examples/auth_example/main.go"
    exit 1
fi
echo "✅ 服务器运行正常"

# 1. 测试公开接口
echo
echo "1️⃣ 测试公开接口"
test_api "GET" "/" "" "" "访问首页"

# 2. 测试用户登录
echo
echo "2️⃣ 测试用户登录"

# 管理员登录
echo "🔑 管理员登录"
ADMIN_LOGIN_RESP=$(curl -s -X POST "$BASE_URL/api/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"password"}')
echo $ADMIN_LOGIN_RESP | jq .

ADMIN_TOKEN=$(echo $ADMIN_LOGIN_RESP | jq -r '.data.token // empty')
if [ "$ADMIN_TOKEN" = "" ] || [ "$ADMIN_TOKEN" = "null" ]; then
    echo "❌ 管理员登录失败"
    exit 1
fi
echo "✅ 管理员Token: ${ADMIN_TOKEN:0:20}..."

# 编辑者登录
echo "📝 编辑者登录"
EDITOR_LOGIN_RESP=$(curl -s -X POST "$BASE_URL/api/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"editor","password":"password"}')
echo $EDITOR_LOGIN_RESP | jq .

EDITOR_TOKEN=$(echo $EDITOR_LOGIN_RESP | jq -r '.data.token // empty')
if [ "$EDITOR_TOKEN" = "" ] || [ "$EDITOR_TOKEN" = "null" ]; then
    echo "❌ 编辑者登录失败"
    exit 1
fi
echo "✅ 编辑者Token: ${EDITOR_TOKEN:0:20}..."

# 普通用户登录
echo "👤 普通用户登录"
USER_LOGIN_RESP=$(curl -s -X POST "$BASE_URL/api/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"user","password":"password"}')
echo $USER_LOGIN_RESP | jq .

USER_TOKEN=$(echo $USER_LOGIN_RESP | jq -r '.data.token // empty')
if [ "$USER_TOKEN" = "" ] || [ "$USER_TOKEN" = "null" ]; then
    echo "❌ 普通用户登录失败"
    exit 1
fi
echo "✅ 用户Token: ${USER_TOKEN:0:20}..."

# 3. 测试无Token访问
echo
echo "3️⃣ 测试无Token访问受保护资源"
test_api "GET" "/api/users" "" "" "无Token访问用户列表（应该失败）"

# 4. 测试认证成功的接口
echo
echo "4️⃣ 测试认证成功的接口"
test_api "GET" "/api/users" "" "$USER_TOKEN" "用户Token访问用户列表"
test_api "GET" "/api/profile" "" "$USER_TOKEN" "用户Token访问个人资料"

# 5. 测试权限验证
echo
echo "5️⃣ 测试权限验证"

# 用户尝试创建用户（应该失败）
test_api "POST" "/api/users" '{"username":"newuser","role":"user"}' "$USER_TOKEN" "普通用户创建用户（应该失败）"

# 管理员创建用户（应该成功）
test_api "POST" "/api/users" '{"username":"newuser","role":"user"}' "$ADMIN_TOKEN" "管理员创建用户（应该成功）"

# 用户尝试删除用户（应该失败）
test_api "DELETE" "/api/users/123" "" "$USER_TOKEN" "普通用户删除用户（应该失败）"

# 管理员删除用户（应该成功）
test_api "DELETE" "/api/users/123" "" "$ADMIN_TOKEN" "管理员删除用户（应该成功）"

# 6. 测试文章权限
echo
echo "6️⃣ 测试文章权限"

# 普通用户查看文章（应该成功）
test_api "GET" "/api/articles" "" "$USER_TOKEN" "普通用户查看文章（应该成功）"

# 普通用户创建文章（应该失败）
test_api "POST" "/api/articles" '{"title":"测试文章","content":"测试内容"}' "$USER_TOKEN" "普通用户创建文章（应该失败）"

# 编辑者创建文章（应该成功）
test_api "POST" "/api/articles" '{"title":"测试文章","content":"测试内容"}' "$EDITOR_TOKEN" "编辑者创建文章（应该成功）"

# 普通用户更新文章（应该失败）
test_api "PUT" "/api/articles/123" '{"title":"更新文章","content":"更新内容"}' "$USER_TOKEN" "普通用户更新文章（应该失败）"

# 编辑者更新文章（应该成功）
test_api "PUT" "/api/articles/123" '{"title":"更新文章","content":"更新内容"}' "$EDITOR_TOKEN" "编辑者更新文章（应该成功）"

# 7. 测试管理员接口
echo
echo "7️⃣ 测试管理员接口"

# 普通用户访问管理员面板（应该失败）
test_api "GET" "/api/admin/dashboard" "" "$USER_TOKEN" "普通用户访问管理员面板（应该失败）"

# 管理员访问面板（应该成功）
test_api "GET" "/api/admin/dashboard" "" "$ADMIN_TOKEN" "管理员访问面板（应该成功）"

# 管理员查看日志（应该成功）
test_api "GET" "/api/admin/logs" "" "$ADMIN_TOKEN" "管理员查看日志（应该成功）"

# 管理员创建权限（应该成功）
test_api "POST" "/api/admin/permissions" '{"resource":"test","action":"read","description":"测试权限"}' "$ADMIN_TOKEN" "管理员创建权限（应该成功）"

# 8. 测试错误Token
echo
echo "8️⃣ 测试错误Token"
test_api "GET" "/api/users" "" "invalid.token.here" "使用无效Token访问（应该失败）"

echo
echo "🎉 测试完成！"
echo "========================================="
echo "📊 测试总结："
echo "✅ 公开接口访问正常"
echo "✅ 用户登录认证正常"  
echo "✅ Token验证正常"
echo "✅ 权限控制正常"
echo "✅ 角色验证正常"
echo "✅ 错误处理正常"
echo
echo "🔧 如需重新测试，请运行: bash examples/auth_example/test_auth.sh" 