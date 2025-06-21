/*
 Navicat Premium Dump SQL

 Source Server         : localhost
 Source Server Type    : MySQL
 Source Server Version : 80037 (8.0.37)
 Source Host           : localhost:3306
 Source Schema         : medical

 Target Server Type    : MySQL
 Target Server Version : 80037 (8.0.37)
 File Encoding         : 65001

 Date: 16/06/2025 15:22:02
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for device
-- ----------------------------
DROP TABLE IF EXISTS `device`;
CREATE TABLE `device`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '设备ID，自增主键',
  `mac_address` varchar(17) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '蓝牙MAC地址，格式如：00:1A:7D:DA:71:13',
  `device_name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '设备名称',
  `device_model` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '设备型号',
  `manufacturer` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '生产厂商',
  `purchase_date` date NULL DEFAULT NULL COMMENT '购买日期',
  `warranty_period` int NULL DEFAULT NULL COMMENT '保修期(月)',
  `status` tinyint(1) NULL DEFAULT 1 COMMENT '设备状态：0-停用，1-启用',
  `last_active_time` datetime NULL DEFAULT NULL COMMENT '最后活跃时间',
  `battery_level` int NULL DEFAULT NULL COMMENT '电池电量百分比',
  `firmware_version` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '固件版本',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_mac_address`(`mac_address` ASC) USING BTREE COMMENT 'MAC地址唯一索引'
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '蓝牙体温计设备表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of device
-- ----------------------------

-- ----------------------------
-- Table structure for device_binding
-- ----------------------------
DROP TABLE IF EXISTS `device_binding`;
CREATE TABLE `device_binding`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '绑定ID，自增主键',
  `device_id` int NOT NULL COMMENT '设备ID',
  `patient_id` int NOT NULL COMMENT '病人ID',
  `binding_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '绑定时间',
  `unbinding_time` datetime NULL DEFAULT NULL COMMENT '解绑时间',
  `is_active` tinyint(1) NULL DEFAULT 1 COMMENT '是否有效：0-无效(已解绑)，1-有效',
  `operator_id` int NULL DEFAULT NULL COMMENT '操作人员ID',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_device_id`(`device_id` ASC) USING BTREE COMMENT '设备ID索引',
  INDEX `idx_patient_id`(`patient_id` ASC) USING BTREE COMMENT '病人ID索引',
  INDEX `idx_is_active`(`is_active` ASC) USING BTREE COMMENT '状态索引'
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '设备与病人绑定关系表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of device_binding
-- ----------------------------

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
) ENGINE = InnoDB AUTO_INCREMENT = 56 CHARACTER SET = utf8mb3 COLLATE = utf8mb3_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of login
-- ----------------------------
INSERT INTO `login` VALUES ('系统管理员', '207cf410532f92a47dee245ce9b11ff71f578ebd763eb3bbea44ebd043d018fb', NULL, NULL, 1, '管理员', 1, 'system', 1, NULL, 0);
INSERT INTO `login` VALUES ('测试用户', '207cf410532f92a47dee245ce9b11ff71f578ebd763eb3bbea44ebd043d018fb', NULL, NULL, 42, NULL, 41, 'test', 1, NULL, 0);

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
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 413 CHARACTER SET = utf8mb3 COLLATE = utf8mb3_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of nav_menu
-- ----------------------------
INSERT INTO `nav_menu` VALUES (34, 0, '系统管理', NULL, 'Baseset', '/manger', 'el-icon-s-help', 1, 2, 0, '菜单', 0);
INSERT INTO `nav_menu` VALUES (38, 34, '用户管理', NULL, 'User_list', '/sys/user', 'el-icon-receiving', 1, 1, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (39, 34, '角色管理', NULL, 'Qxzlist', '/sys/role', 'el-icon-collection', 1, 2, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (41, 34, '菜单管理', NULL, 'Nav', '/sys/resource/menu', 'el-icon-files', 1, 3, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (86, 34, '数据字典', NULL, 'Valdict', '/dev/dict', 'el-icon-toilet-paper', 1, 5, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (166, 34, '清除缓存', NULL, 'clearcache', '/manger/clearCatch', 'el-icon-office-building', 0, 4, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (219, 34, '路由设置', NULL, 'routesetup', '/manger/routemap', 'el-icon-suitcase', 1, 50, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (226, 34, '路由权限', NULL, 'routeqxguanli', '/manger/routerole', 'component', 0, 8, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (344, 0, '数据看板', NULL, 'reportlist', '/Report', 'table', 1, 0, 0, '菜单', 0);
INSERT INTO `nav_menu` VALUES (352, 206, 'My workplace', NULL, 'My workplace', '/Application/workplace', 'el-icon-guide', 1, 40, 0, '功能按钮', 0);
INSERT INTO `nav_menu` VALUES (353, 344, '数据看板', NULL, 'databasekangban', '/Report/dataview', 'el-icon-guide', 1, 5, 0, '功能按钮', 0);
INSERT INTO `nav_menu` VALUES (354, 34, '产线管理', NULL, NULL, '/bs/sl_department', 'el-icon-cpu', 1, 3, 0, NULL, 0);
INSERT INTO `nav_menu` VALUES (355, 392, '产线上料工位设置', NULL, NULL, '/bs/sl_station', 'el-icon-connection', 1, 3, 0, NULL, 0);
INSERT INTO `nav_menu` VALUES (367, 83, '工站扩展属性录入', 'station_label Input', 'sl_station_ext', '/bs/sl_station_ext', 'el-icon-s-grid', 1, 9, NULL, NULL, 0);
INSERT INTO `nav_menu` VALUES (368, 34, '系统版本', 'system version', 'sysver', '/bs/sysver', 'el-icon-umbrella', 1, 7, NULL, NULL, 0);
INSERT INTO `nav_menu` VALUES (392, 0, '库位管理', NULL, 'station', '/station', 'el-icon-s-management', 1, 1, 0, '菜单', 0);
INSERT INTO `nav_menu` VALUES (394, 392, '物料存放类型设置', NULL, 'store', '/station/store', 'el-icon-s-opportunity\r\n', 1, 0, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (395, 392, '库位管理', '超市存放点设置', 'warehouse', '/station/warehouse', 'el-icon-present', 1, 2, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (396, 392, '任务查看', 'Task', 'task', '/station/task', 'el-icon-reading', 1, 10, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (397, 392, '警报查看', 'Alarm', 'alarm', '/station/alarm', 'el-icon-sunrise', 1, 10, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (398, 392, '日志管理', NULL, 'log', '/manger/log', 'el-icon-pie-chart', 1, 10, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (400, 392, '原料BOM', NULL, 'raw_material', '/station/raw_material', 'el-icon-s-open', 1, 1, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (401, 392, '上料工位对应原材料型号设置', NULL, 'station_list', '/station/station_list', 'el-icon-s-order', 0, 3, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (402, 392, 'AGV点位设置', NULL, 'agvcode', '/station/agvcode', 'el-icon-c-scale-to-original', 0, 6, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (403, 344, '大屏数据', 'Task Map', 'Task Map', '/bigdata', 'el-icon-s-flag', 1, NULL, 0, NULL, 0);
INSERT INTO `nav_menu` VALUES (404, 0, '个人中心', 'usercenter', 'usercenter', '/profile/index', 'el-icon-s-flag', 1, NULL, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (407, 392, '生产型号管理', 'production', 'production', '/station/production', 'el-icon-s-cooperation', 1, 67, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (408, 392, '传感器状态', 'device_status', 'device_status', '/station/device_status', 'el-icon-s-management', 1, 68, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (409, 392, '库位状态', 'store_queue', 'store_queue', '/station/store_queue', 'el-icon-date', 1, 69, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (410, 34, '联系人设置', 'contacts', 'contacts', '/bs/contacts', 'el-icon-phone', 1, 7, 0, '页面', 0);
INSERT INTO `nav_menu` VALUES (411, 396, '取消任务', 'task_cancel', 'task_cancel', '/station/task/cancel', 'el-icon-s-management', 1, 0, 0, '按钮', 0);
INSERT INTO `nav_menu` VALUES (412, 344, '地图编辑', 'map_edit', 'map_edit', '/mapedit', 'el-icon-s-flag', 1, NULL, 0, NULL, 0);

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
) ENGINE = InnoDB AUTO_INCREMENT = 45 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '系统访问日志表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of operation_log
-- ----------------------------
INSERT INTO `operation_log` VALUES (1, '::1', '/swagger', '2025-06-10 16:55:07', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (2, '::1', '/swagger/openapi.json', '2025-06-10 16:55:07', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (3, '::1', '/api/auth/userinfo', '2025-06-10 16:55:18', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (4, '::1', '/api/auth/login', '2025-06-10 16:56:29', '0', 'POST', '', '{\r\n    \"code\": \"system\",\r\n    \"password\": \"feaa38b7b97cc01e7c95554a48b2b8739f119ba2fc9bfc4ed6bf53e20aae10aef450e95c5d2ebbb4d9b8397fb2d3802cfc066164ed5bc26753cee6484c80ef645a2c2238eeea8f15188353401b557ae58358bb42edc5dda6a2559947f2977a6d6f0ca93b5d43\"\r\n}');
INSERT INTO `operation_log` VALUES (5, '::1', '/api/db/detail', '2025-06-10 16:58:18', '0', 'GET', 'id=01', '');
INSERT INTO `operation_log` VALUES (6, '::1', '/api/auth/userinfo', '2025-06-10 17:01:22', '1', 'GET', '', '');
INSERT INTO `operation_log` VALUES (7, '::1', '/swagger', '2025-06-10 17:19:02', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (8, '::1', '/swagger/openapi.json', '2025-06-10 17:19:03', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (9, '::1', '/swagger', '2025-06-11 08:25:45', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (10, '::1', '/swagger/openapi.json', '2025-06-11 08:25:45', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (11, '::1', '/swagger', '2025-06-11 08:31:51', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (12, '::1', '/swagger/openapi.json', '2025-06-11 08:31:51', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (13, '::1', '/swagger', '2025-06-11 08:38:39', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (14, '::1', '/swagger/openapi.json', '2025-06-11 08:38:40', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (15, '::1', '/swagger', '2025-06-11 08:42:42', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (16, '::1', '/swagger/openapi.json', '2025-06-11 08:42:43', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (17, '::1', '/swagger', '2025-06-11 08:45:48', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (18, '::1', '/swagger/openapi.json', '2025-06-11 08:45:48', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (19, '::1', '/swagger', '2025-06-11 08:53:10', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (20, '::1', '/swagger/openapi.json', '2025-06-11 08:53:10', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (21, '::1', '/swagger', '2025-06-11 08:56:20', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (22, '::1', '/swagger/openapi.json', '2025-06-11 08:56:21', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (23, '::1', '/swagger', '2025-06-11 09:17:35', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (24, '::1', '/swagger/openapi.json', '2025-06-11 09:17:35', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (25, '::1', '/swagger', '2025-06-11 09:28:33', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (26, '::1', '/swagger/openapi.json', '2025-06-11 09:28:33', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (27, '::1', '/swagger', '2025-06-11 09:32:24', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (28, '::1', '/swagger/openapi.json', '2025-06-11 09:32:24', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (29, '::1', '/swagger', '2025-06-11 09:36:38', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (30, '::1', '/swagger/openapi.json', '2025-06-11 09:36:38', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (31, '::1', '/swagger', '2025-06-11 09:38:47', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (32, '::1', '/swagger/openapi.json', '2025-06-11 09:38:47', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (33, '::1', '/swagger', '2025-06-11 17:08:41', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (34, '::1', '/swagger/openapi.json', '2025-06-11 17:08:42', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (35, '::1', '/swagger', '2025-06-11 17:15:42', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (36, '::1', '/swagger/openapi.json', '2025-06-11 17:15:43', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (37, '::1', '/swagger', '2025-06-11 17:20:30', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (38, '::1', '/swagger/openapi.json', '2025-06-11 17:20:30', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (39, '::1', '/swagger', '2025-06-16 14:26:25', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (40, '::1', '/swagger/openapi.json', '2025-06-16 14:26:25', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (41, '::1', '/swagger/openapi.json', '2025-06-16 14:34:26', '0', 'GET', '', '');
INSERT INTO `operation_log` VALUES (42, '::1', '/api/auth/login', '2025-06-16 14:38:59', '0', 'POST', '', '{\r\n    \"code\": \"system\",\r\n    \"password\": \"tempor deserunt\"\r\n}');
INSERT INTO `operation_log` VALUES (43, '::1', '/api/auth/login', '2025-06-16 14:43:08', '0', 'POST', '', '{\r\n    \"code\": \"system\",\r\n    \"password\": \"tempor deserunt\"\r\n}');
INSERT INTO `operation_log` VALUES (44, '::1', '/api/auth/getmenu', '2025-06-16 14:54:32', '0', 'GET', '', '');

-- ----------------------------
-- Table structure for patient
-- ----------------------------
DROP TABLE IF EXISTS `patient`;
CREATE TABLE `patient`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '病人ID，自增主键',
  `patient_no` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '病人编号',
  `name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '病人姓名',
  `gender` tinyint(1) NULL DEFAULT NULL COMMENT '性别：0-女，1-男',
  `birth_date` date NULL DEFAULT NULL COMMENT '出生日期',
  `age` int NULL DEFAULT NULL COMMENT '年龄',
  `phone` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '联系电话',
  `address` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '住址',
  `id_card` varchar(18) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '身份证号',
  `medical_history` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '病史记录',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_patient_no`(`patient_no` ASC) USING BTREE COMMENT '病人编号唯一索引'
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '病人信息表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of patient
-- ----------------------------

-- ----------------------------
-- Table structure for posture_analysis
-- ----------------------------
DROP TABLE IF EXISTS `posture_analysis`;
CREATE TABLE `posture_analysis`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '分析ID，自增主键',
  `record_id` int NOT NULL COMMENT '关联的传感器记录ID',
  `patient_id` int NOT NULL COMMENT '病人ID',
  `posture_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '姿态类型：平躺/侧卧/坐立/站立/运动等',
  `posture_confidence` decimal(3, 2) NULL DEFAULT NULL COMMENT '姿态识别置信度(0-1)',
  `movement_intensity` decimal(4, 2) NULL DEFAULT NULL COMMENT '运动强度',
  `stability_score` decimal(3, 2) NULL DEFAULT NULL COMMENT '稳定性评分(0-1)',
  `analysis_time` datetime NOT NULL COMMENT '分析时间',
  `analysis_status` tinyint(1) NULL DEFAULT 1 COMMENT '分析状态：0-失败，1-成功',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_record_id`(`record_id` ASC) USING BTREE COMMENT '记录ID索引',
  INDEX `idx_patient_id`(`patient_id` ASC) USING BTREE COMMENT '病人ID索引',
  INDEX `idx_posture_type`(`posture_type` ASC) USING BTREE COMMENT '姿态类型索引'
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '姿态分析记录表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of posture_analysis
-- ----------------------------

-- ----------------------------
-- Table structure for role
-- ----------------------------
DROP TABLE IF EXISTS `role`;
CREATE TABLE `role`  (
  `name` varchar(50) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NOT NULL COMMENT '角色名称',
  `memo` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL COMMENT '授权内容',
  `id` int NOT NULL AUTO_INCREMENT,
  `is_default` int NULL DEFAULT 0,
  `memoview` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NULL COMMENT ' webfrontdata',
  `px` int NULL DEFAULT NULL,
  `is_del` int NULL DEFAULT 0,
  `idm_id` int NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 49 CHARACTER SET = utf8mb3 COLLATE = utf8mb3_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of role
-- ----------------------------
INSERT INTO `role` VALUES ('超级管理', '404,344,403,412,353,392,394,400,395,355,396,411,397,398,407,408,409,34,38,39,41,354,86,368,410,219', 1, 0, '404,344,403,412,353,392,394,400,395,355,396,411,397,398,407,408,409,34,38,39,41,354,86,368,410,219', 12, 0, 0);
INSERT INTO `role` VALUES ('Guest', '404,344,403,353', 42, 1, '404,344,403,353', 50, 0, 0);

-- ----------------------------
-- Table structure for sensor_record
-- ----------------------------
DROP TABLE IF EXISTS `sensor_record`;
CREATE TABLE `sensor_record`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '记录ID，自增主键',
  `device_id` int NOT NULL COMMENT '设备ID',
  `patient_id` int NOT NULL COMMENT '病人ID',
  `temperature` decimal(4, 2) NOT NULL COMMENT '体温值(摄氏度)',
  `accel_x` decimal(6, 3) NULL DEFAULT NULL COMMENT '加速度X轴(m/s²)',
  `accel_y` decimal(6, 3) NULL DEFAULT NULL COMMENT '加速度Y轴(m/s²)',
  `accel_z` decimal(6, 3) NULL DEFAULT NULL COMMENT '加速度Z轴(m/s²)',
  `gyro_x` decimal(6, 3) NULL DEFAULT NULL COMMENT '陀螺仪X轴(°/s)',
  `gyro_y` decimal(6, 3) NULL DEFAULT NULL COMMENT '陀螺仪Y轴(°/s)',
  `gyro_z` decimal(6, 3) NULL DEFAULT NULL COMMENT '陀螺仪Z轴(°/s)',
  `battery_level` int NULL DEFAULT NULL COMMENT '电量',
  `measure_time` datetime NOT NULL COMMENT '测量时间',
  `measure_location` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '测量部位(腋下/口腔/肛门/耳温等)',
  `measure_status` tinyint(1) NULL DEFAULT 1 COMMENT '测量状态：0-异常，1-正常',
  `data_quality` tinyint(1) NULL DEFAULT 1 COMMENT '数据质量：0-低质量，1-高质量',
  `mqtt_topic` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'MQTT主题',
  `mqtt_message_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'MQTT消息ID',
  `raw_data` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '原始16进制数据',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_device_id`(`device_id` ASC) USING BTREE COMMENT '设备ID索引',
  INDEX `idx_patient_id`(`patient_id` ASC) USING BTREE COMMENT '病人ID索引',
  INDEX `idx_measure_time`(`measure_time` ASC) USING BTREE COMMENT '测量时间索引',
  INDEX `idx_temperature`(`temperature` ASC) USING BTREE COMMENT '体温索引'
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '传感器测量记录表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of sensor_record
-- ----------------------------

-- ----------------------------
-- Table structure for temperature_alert
-- ----------------------------
DROP TABLE IF EXISTS `temperature_alert`;
CREATE TABLE `temperature_alert`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '报警ID，自增主键',
  `record_id` int NOT NULL COMMENT '关联的传感器记录ID',
  `patient_id` int NOT NULL COMMENT '病人ID',
  `temperature` decimal(4, 2) NOT NULL COMMENT '报警体温值',
  `alert_type` tinyint(1) NOT NULL COMMENT '报警类型：1-低烧，2-高烧，3-超高烧',
  `alert_time` datetime NOT NULL COMMENT '报警时间',
  `alert_status` tinyint(1) NULL DEFAULT 0 COMMENT '处理状态：0-未处理，1-已处理',
  `handler_id` int NULL DEFAULT NULL COMMENT '处理人员ID',
  `handle_time` datetime NULL DEFAULT NULL COMMENT '处理时间',
  `handle_notes` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '处理备注',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_record_id`(`record_id` ASC) USING BTREE COMMENT '记录ID索引',
  INDEX `idx_patient_id`(`patient_id` ASC) USING BTREE COMMENT '病人ID索引',
  INDEX `idx_alert_status`(`alert_status` ASC) USING BTREE COMMENT '报警状态索引'
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '体温异常报警记录表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of temperature_alert
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
