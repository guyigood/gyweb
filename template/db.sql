/*
 Navicat Premium Dump SQL

 Source Server         : localhost
 Source Server Type    : MySQL
 Source Server Version : 80037 (8.0.37)
 Source Host           : localhost:3306
 Source Schema         : ejkcrm

 Target Server Type    : MySQL
 Target Server Version : 80037 (8.0.37)
 File Encoding         : 65001

 Date: 21/07/2025 08:30:17
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for login
-- ----------------------------
DROP TABLE IF EXISTS `login`;
CREATE TABLE `login`  (
  `name` varchar(50) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL DEFAULT NULL COMMENT '用户名',
  `pass` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL DEFAULT '123456' COMMENT '密码',
  `email` varchar(200) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL DEFAULT NULL COMMENT '邮箱',
  `memo` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL COMMENT '备注',
  `role_id` int NULL DEFAULT NULL,
  `tel` varchar(50) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL DEFAULT NULL COMMENT '联系电话',
  `id` int NOT NULL AUTO_INCREMENT,
  `code` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL DEFAULT NULL COMMENT '登陆代码',
  `status` int NULL DEFAULT 1 COMMENT '状态（1有效，2禁用）',
  `head_pic` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL DEFAULT NULL COMMENT '头像',
  `is_del` int NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 59 CHARACTER SET = utf8mb3 COLLATE = utf8mb3_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of login
-- ----------------------------
INSERT INTO `login` VALUES ('系统管理员', '0192023a7bbd73250516f069df18b500', NULL, '207cf410532f92a47dee245ce9b11ff71f578ebd763eb3bbea44ebd043d018fb', 1, '管理员', 1, 'admin', 1, NULL, 0);
INSERT INTO `login` VALUES ('test', 'e10adc3949ba59abbe56e057f20f883e', NULL, '207cf410532f92a47dee245ce9b11ff71f578ebd763eb3bbea44ebd043d018fb', 42, NULL, 41, 'test', 1, NULL, 0);
INSERT INTO `login` VALUES ('32234', '123456', '双方发生的', NULL, 50, '的撒发射点发', 58, NULL, 1, NULL, 1);

-- ----------------------------
-- Table structure for nav_menu
-- ----------------------------
DROP TABLE IF EXISTS `nav_menu`;
CREATE TABLE `nav_menu`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `parent_id` int NULL DEFAULT NULL COMMENT '父级菜单',
  `nav_name` varchar(200) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL DEFAULT NULL COMMENT '菜单名称',
  `en_name` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL DEFAULT NULL COMMENT '英文名称',
  `nav_code` varchar(200) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL DEFAULT NULL COMMENT '菜单代码',
  `nav_module` varchar(200) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL DEFAULT NULL COMMENT '菜单模块',
  `nav_image` varchar(300) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL DEFAULT NULL COMMENT '菜单图像',
  `is_display` int NULL DEFAULT 0,
  `order_number` int NULL DEFAULT NULL COMMENT '显示顺序',
  `is_tel` int NULL DEFAULT 0,
  `nav_type` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL DEFAULT NULL COMMENT '功能属性',
  `is_del` int NULL DEFAULT 0,
  `path` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 26 CHARACTER SET = utf8mb3 COLLATE = utf8mb3_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of nav_menu
-- ----------------------------
INSERT INTO `nav_menu` VALUES (1, 0, '医疗管理系统', 'Medical', 'medical', 'medical', 'ri:hospital-line', 1, 0, 0, 'menu', 0, '/');
INSERT INTO `nav_menu` VALUES (2, 25, '医疗仪表盘', 'MedicalDashboard', 'medical_dashboard', 'medical', 'ep:data-analysis', 0, 1, 0, 'page', 0, '/medical/dashboard');
INSERT INTO `nav_menu` VALUES (3, 25, '设备管理', 'DeviceManagement', 'device_management', 'medical', 'ep:cpu', 1, 2, 0, 'menu', 1, '/medical/devices/dlist');
INSERT INTO `nav_menu` VALUES (4, 25, '病人管理', 'PatientManagement', 'patient_management', 'medical', 'ep:user', 1, 3, 0, 'menu', 1, '/medical/patients/list');
INSERT INTO `nav_menu` VALUES (5, 25, '设备绑定管理', 'DeviceBinding', 'device_binding', 'medical', 'ep:link', 1, 4, 0, 'page', 0, '/medical/bindings');
INSERT INTO `nav_menu` VALUES (6, 1, '实时监控', 'MonitoringManagement', 'medical_monitoring', 'medical', 'ep:monitor', 1, 5, 0, 'menu', 0, '/medical/monitoring');
INSERT INTO `nav_menu` VALUES (7, 6, '报警管理', 'AlertManagement', 'alert_management', 'medical', 'ep:warning', 1, 6, 0, 'page', 0, '/medical/alerts');
INSERT INTO `nav_menu` VALUES (8, 1, '数据统计', 'StatisticsManagement', 'medical_statistics', 'medical', 'ep:data-analysis', 0, 7, 0, 'menu', 0, '/medical/statistics');
INSERT INTO `nav_menu` VALUES (9, 1, '系统管理', 'SystemManagement', 'medical_system', 'medical', 'ep:setting', 1, 8, 0, 'menu', 0, '/medical/system');
INSERT INTO `nav_menu` VALUES (10, 25, '设备列表', 'DeviceList', 'device_list', 'medical', 'ep:list', 1, 1, 0, 'page', 0, '/medical/devices/list');
INSERT INTO `nav_menu` VALUES (13, 25, '病人列表', 'PatientList', 'patient_list', 'medical', 'ep:avatar', 1, 1, 0, 'page', 0, '/medical/patients/list');
INSERT INTO `nav_menu` VALUES (16, 6, '实时数据监控', 'RealTimeData', 'realtime_data', 'medical', 'ep:data-line', 0, 1, 0, 'page', 0, '/medical/monitoring/realtime');
INSERT INTO `nav_menu` VALUES (17, 6, '体温记录查看', 'TemperatureRecords', 'temperature_records', 'medical', 'ep:list', 1, 2, 0, 'page', 0, '/medical/monitoring/temperature');
INSERT INTO `nav_menu` VALUES (18, 8, '统计概览', 'StatisticsOverview', 'statistics_overview', 'medical', 'ep:pie-chart', 1, 1, 0, 'page', 0, '/medical/statistics/overview');
INSERT INTO `nav_menu` VALUES (19, 8, '报表管理', 'Reports', 'report_management', 'medical', 'ep:document', 0, 2, 0, 'page', 0, '/medical/statistics/reports');
INSERT INTO `nav_menu` VALUES (20, 9, '用户管理', 'UserManagement', 'user_management', 'medical', 'ep:user-filled', 1, 1, 0, 'page', 0, '/medical/system/users');
INSERT INTO `nav_menu` VALUES (21, 9, '角色设置', 'RoleSettings', 'role_settings', 'medical', 'ep:avatar', 1, 2, 0, 'page', 0, '/medical/system/roles');
INSERT INTO `nav_menu` VALUES (22, 9, '菜单设置', 'MenuSettings', 'menu_settings', 'medical', 'ep:menu', 1, 3, 0, 'page', 0, '/medical/system/menus');
INSERT INTO `nav_menu` VALUES (23, 9, '数据模块设置', 'DataSettings', 'data_settings', 'medical', 'ep:tools', 1, 4, 0, 'page', 0, '/medical/system/settings');
INSERT INTO `nav_menu` VALUES (24, 9, '系统日志', 'SystemLogs', 'system_logs', 'medical', 'ep:document', 1, 5, 0, 'page', 0, '/medical/system/logs');
INSERT INTO `nav_menu` VALUES (25, 1, '业务管理', 'Business', 'business', 'medical', 'ep:box', 1, 1, 0, 'menu', 0, '/medical/business');

-- ----------------------------
-- Table structure for operation_log
-- ----------------------------
DROP TABLE IF EXISTS `operation_log`;
CREATE TABLE `operation_log`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `ip` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '客户端IP地址',
  `url` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '请求URL路径',
  `add_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  `user_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '用户ID（可为空）',
  `method` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'HTTP请求方法',
  `params` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '请求参数',
  `body` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '请求体内容',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 6475 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '系统访问日志表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of operation_log
-- ----------------------------

-- ----------------------------
-- Table structure for role
-- ----------------------------
DROP TABLE IF EXISTS `role`;
CREATE TABLE `role`  (
  `name` varchar(50) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NOT NULL COMMENT '角色名称',
  `memo` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL COMMENT '授权内容',
  `id` int NOT NULL AUTO_INCREMENT,
  `is_default` int NULL DEFAULT 0 COMMENT '是否缺省',
  `px` int NULL DEFAULT NULL,
  `is_del` int NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 51 CHARACTER SET = utf8mb3 COLLATE = utf8mb3_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of role
-- ----------------------------
INSERT INTO `role` VALUES ('超级管理', '1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25', 1, 0, 12, 0);
INSERT INTO `role` VALUES ('Guest', '1,8,18,19', 42, 1, 50, 0);
INSERT INTO `role` VALUES ('一般权限', '1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23', 49, 0, 3, 0);
INSERT INTO `role` VALUES ('医生', '2,3,4,5,6,7,8,9,10,12,13,15,16,17,18,19,20,21,22,23', 50, 0, 5, 0);

-- ----------------------------
-- Table structure for sys_field_info
-- ----------------------------
DROP TABLE IF EXISTS `sys_field_info`;
CREATE TABLE `sys_field_info`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `table_id` int NOT NULL COMMENT '所属表ID(关联sys_table_info.id)',
  `field_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '字段名',
  `field_comment` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '字段注释',
  `field_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '字段类型(varchar/int/datetime等)',
  `field_length` int NULL DEFAULT NULL COMMENT '字段长度',
  `default_value` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '默认值',
  `is_pk` tinyint(1) NULL DEFAULT 0 COMMENT '是否主键(0:否 1:是)',
  `is_required` tinyint(1) NULL DEFAULT 0 COMMENT '是否必填(0:否 1:是)',
  `is_unique` tinyint(1) NULL DEFAULT 0 COMMENT '是否唯一(0:否 1:是)',
  `show_in_list` tinyint(1) NULL DEFAULT 1 COMMENT '列表中显示(0:否 1:是)',
  `show_in_add` tinyint(1) NULL DEFAULT 1 COMMENT '新增时显示(0:否 1:是)',
  `show_in_edit` tinyint(1) NULL DEFAULT 1 COMMENT '编辑时显示(0:否 1:是)',
  `show_in_detail` tinyint(1) NULL DEFAULT 1 COMMENT '详情中显示(0:否 1:是)',
  `readonly_in_edit` tinyint(1) NULL DEFAULT 0 COMMENT '编辑时只读(0:否 1:是)',
  `is_searchable` tinyint(1) NULL DEFAULT 0 COMMENT '是否可查询(0:否 1:是)',
  `query_type` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'EQ' COMMENT '查询类型(EQ/LIKE/IN/BETWEEN等)',
  `query_weight` int NULL DEFAULT 1 COMMENT '查询权重(影响查询条件顺序)精简查询和详细查询的区别',
  `list_sort_order` int NULL DEFAULT 1 COMMENT '列表显示顺序',
  `form_sort_order` int NULL DEFAULT 1 COMMENT '表单显示顺序',
  `query_sort_order` int NULL DEFAULT 1 COMMENT '查询条件显示顺序',
  `form_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'input' COMMENT '表单控件类型(input/select/textarea/date等)',
  `form_size` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'default' COMMENT '控件大小(large/default/small)',
  `placeholder` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '输入提示文本',
  `validation_rule` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '验证规则(JSON格式)',
  `ref_table` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '关联表名',
  `ref_field` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '关联字段',
  `ref_display_field` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '关联显示字段',
  `ref_api_url` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '关联数据API地址',
  `list_width` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '列表列宽(如:120px/10%)',
  `list_align` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'left' COMMENT '列表对齐方式(left/center/right)',
  `list_fixed` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '列表固定方式(left/right)',
  `list_format` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '列表格式化类型(date/money/percent等)',
  `list_format_pattern` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '格式化模式',
  `is_active` tinyint(1) NULL DEFAULT 1 COMMENT '是否启用(0:否 1:是)',
  `group_name` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '字段分组名称',
  `help_text` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '帮助文本',
  `remark` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '备注',
  `create_time` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `create_by` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `update_by` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_table_field`(`table_id` ASC, `field_name` ASC) USING BTREE,
  INDEX `idx_table_id`(`table_id` ASC) USING BTREE,
  INDEX `idx_field_name`(`field_name` ASC) USING BTREE,
  INDEX `idx_is_searchable`(`is_searchable` ASC) USING BTREE,
  INDEX `idx_show_in_list`(`show_in_list` ASC) USING BTREE,
  INDEX `idx_query_type`(`query_type` ASC) USING BTREE,
  INDEX `idx_form_type`(`form_type` ASC) USING BTREE,
  INDEX `idx_field_info_composite`(`table_id` ASC, `is_active` ASC, `show_in_list` ASC) USING BTREE,
  INDEX `idx_field_info_search`(`table_id` ASC, `is_searchable` ASC, `query_sort_order` ASC) USING BTREE,
  INDEX `idx_field_info_form`(`table_id` ASC, `is_active` ASC, `form_sort_order` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 196 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '数据库字段信息配置表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of sys_field_info
-- ----------------------------

-- ----------------------------
-- Table structure for sys_query_type
-- ----------------------------
DROP TABLE IF EXISTS `sys_query_type`;
CREATE TABLE `sys_query_type`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `type_code` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '查询类型代码',
  `type_name` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '查询类型名称',
  `operator_symbol` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '操作符号',
  `description` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '描述',
  `is_multi_value` tinyint(1) NULL DEFAULT 0 COMMENT '是否多值查询(0:否 1:是)',
  `sort_order` int NULL DEFAULT 1 COMMENT '排序',
  `is_active` tinyint(1) NULL DEFAULT 1 COMMENT '是否启用(0:否 1:是)',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_type_code`(`type_code` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 16 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '查询类型字典表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of sys_query_type
-- ----------------------------
INSERT INTO `sys_query_type` VALUES (1, 'EQ', '等于', '=', '精确匹配', 0, 1, 1);
INSERT INTO `sys_query_type` VALUES (2, 'NE', '不等于', '!=', '不等于匹配', 0, 2, 1);
INSERT INTO `sys_query_type` VALUES (3, 'GT', '大于', '>', '大于比较', 0, 3, 1);
INSERT INTO `sys_query_type` VALUES (4, 'GE', '大于等于', '>=', '大于等于比较', 0, 4, 1);
INSERT INTO `sys_query_type` VALUES (5, 'LT', '小于', '<', '小于比较', 0, 5, 1);
INSERT INTO `sys_query_type` VALUES (6, 'LE', '小于等于', '<=', '小于等于比较', 0, 6, 1);
INSERT INTO `sys_query_type` VALUES (7, 'LIKE', '模糊查询', 'LIKE', '模糊匹配', 0, 7, 1);
INSERT INTO `sys_query_type` VALUES (8, 'LEFT_LIKE', '左模糊', 'LIKE', '左模糊匹配(%关键字)', 0, 8, 1);
INSERT INTO `sys_query_type` VALUES (9, 'RIGHT_LIKE', '右模糊', 'LIKE', '右模糊匹配(关键字%)', 0, 9, 1);
INSERT INTO `sys_query_type` VALUES (10, 'IN', '包含', 'IN', '在指定值列表中', 1, 10, 1);
INSERT INTO `sys_query_type` VALUES (11, 'NOT_IN', '不包含', 'NOT IN', '不在指定值列表中', 1, 11, 1);
INSERT INTO `sys_query_type` VALUES (12, 'BETWEEN', '区间查询', 'BETWEEN', '在指定范围内', 1, 12, 1);
INSERT INTO `sys_query_type` VALUES (13, 'IS_NULL', '为空', 'IS NULL', '字段值为NULL', 0, 13, 1);
INSERT INTO `sys_query_type` VALUES (14, 'IS_NOT_NULL', '不为空', 'IS NOT NULL', '字段值不为NULL', 0, 14, 1);
INSERT INTO `sys_query_type` VALUES (15, 'REGEXP', '正则匹配', 'REGEXP', '正则表达式匹配', 0, 15, 1);

-- ----------------------------
-- Table structure for sys_table_info
-- ----------------------------
DROP TABLE IF EXISTS `sys_table_info`;
CREATE TABLE `sys_table_info`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `table_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '表名',
  `table_comment` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '表注释',
  `module_name` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '所属模块',
  `pk_field` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'id' COMMENT '主键字段名',
  `sort_field` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '默认排序字段',
  `sort_order` varchar(4) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'ASC' COMMENT '排序方式(ASC/DESC)',
  `page_size` int NULL DEFAULT 20 COMMENT '分页大小',
  `enable_export` tinyint(1) NULL DEFAULT 1 COMMENT '是否支持导出(0:否 1:是)',
  `enable_import` tinyint(1) NULL DEFAULT 0 COMMENT '是否支持导入(0:否 1:是)',
  `enable_batch_delete` tinyint(1) NULL DEFAULT 1 COMMENT '是否支持批量删除(0:否 1:是)',
  `join_tables` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '关联表配置(JSON格式存储多个关联表信息)',
  `join_field_alias` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '关联字段别名',
  `list_api_url` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '列表查询API地址',
  `detail_api_url` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '详情查询API地址',
  `add_api_url` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '新增API地址',
  `edit_api_url` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '编辑API地址',
  `delete_api_url` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '删除API地址',
  `status` tinyint(1) NULL DEFAULT 1 COMMENT '状态(0:禁用 1:启用)',
  `remark` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '备注',
  `create_time` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `create_by` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `update_by` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_table_name`(`table_name` ASC) USING BTREE,
  INDEX `idx_module_name`(`module_name` ASC) USING BTREE,
  INDEX `idx_status`(`status` ASC) USING BTREE,
  INDEX `idx_table_info_composite`(`module_name` ASC, `status` ASC, `table_name` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 15 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '数据库表信息配置表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of sys_table_info
-- ----------------------------

-- ----------------------------
-- Table structure for valdict
-- ----------------------------
DROP TABLE IF EXISTS `valdict`;
CREATE TABLE `valdict`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL DEFAULT NULL COMMENT '分类名称',
  `dict_key` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL DEFAULT NULL COMMENT '关键字',
  `dict_value` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL DEFAULT NULL COMMENT '属性值',
  `img` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL DEFAULT NULL COMMENT '图片',
  `px` int NULL DEFAULT NULL COMMENT '排序',
  `is_del` int NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 220 CHARACTER SET = utf8mb3 COLLATE = utf8mb3_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of valdict
-- ----------------------------

SET FOREIGN_KEY_CHECKS = 1;
