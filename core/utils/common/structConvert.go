package common

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// StructToMap 将任意结构体转换为map[string]interface{}
// 支持嵌套结构体、指针、切片、数组等复杂类型
// 参数:
//
//	obj: 要转换的结构体对象，可以是值或指针
//
// 返回值:
//
//	map[string]interface{}: 转换后的map，key为字段名（支持json tag），value为字段值
func StructToMap(obj interface{}) map[string]interface{} {
	if obj == nil {
		return nil
	}

	result := make(map[string]interface{})
	v := reflect.ValueOf(obj)
	t := reflect.TypeOf(obj)

	// 如果是指针，获取指向的值
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
		t = t.Elem()
	}

	// 只处理结构体类型
	if v.Kind() != reflect.Struct {
		return nil
	}

	// 遍历结构体字段
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// 跳过未导出的字段
		if !fieldValue.CanInterface() {
			continue
		}

		// 获取字段名，优先使用json 和 db tag
		fieldName := getFieldName(field)
		if fieldName == "-" {
			continue // 跳过标记为忽略的字段
		}

		// 转换字段值
		result[fieldName] = convertValue(fieldValue)
	}

	return result
}

// MapToStruct 将map[string]interface{}转换为指定的结构体
// 支持嵌套结构体、指针、切片、数组等复杂类型
// 参数:
//
//	m: 源map数据
//	obj: 目标结构体的指针（必须是指针类型）
//
// 返回值:
//
//	error: 转换过程中的错误信息
func MapToStruct(m map[string]interface{}, obj interface{}) error {
	if m == nil {
		return fmt.Errorf("source map is nil")
	}
	if obj == nil {
		return fmt.Errorf("target object is nil")
	}

	v := reflect.ValueOf(obj)
	t := reflect.TypeOf(obj)

	// 必须是指针类型
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("target object must be a pointer")
	}

	// 指针不能为nil
	if v.IsNil() {
		return fmt.Errorf("target object pointer is nil")
	}

	// 获取指针指向的值
	v = v.Elem()
	t = t.Elem()

	// 必须是结构体类型
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("target object must be a struct")
	}

	// 遍历结构体字段
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// 跳过不可设置的字段
		if !fieldValue.CanSet() {
			continue
		}

		// 获取字段名
		fieldName := getFieldName(field)
		if fieldName == "-" {
			continue // 跳过标记为忽略的字段
		}

		// 从map中获取对应的值
		mapValue, exists := m[fieldName]
		if !exists {
			continue // 如果map中没有对应的key，跳过
		}

		// 设置字段值
		if err := setFieldValue(fieldValue, mapValue); err != nil {
			return fmt.Errorf("failed to set field %s: %v", fieldName, err)
		}
	}

	return nil
}

// getFieldName 获取字段名，优先使用json tag
func getFieldName(field reflect.StructField) string {
	// 检查json tag
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		// 处理json tag中的选项（如omitempty）
		if idx := strings.Index(jsonTag, ","); idx != -1 {
			return jsonTag[:idx]
		}
		return jsonTag
	}

	// 检查db tag
	if dbTag := field.Tag.Get("db"); dbTag != "" {
		return dbTag
	}

	// 使用字段名
	return field.Name
}

// convertValue 转换reflect.Value为interface{}
func convertValue(v reflect.Value) interface{} {
	if !v.IsValid() {
		return nil
	}

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return nil
		}
		return convertValue(v.Elem())

	case reflect.Struct:
		// 特殊处理时间类型
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return v.Interface().(time.Time)
		}
		// 递归转换嵌套结构体
		return StructToMap(v.Interface())

	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return []interface{}{}
		}
		result := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = convertValue(v.Index(i))
		}
		return result

	case reflect.Map:
		if v.Len() == 0 {
			return map[string]interface{}{}
		}
		result := make(map[string]interface{})
		for _, key := range v.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			result[keyStr] = convertValue(v.MapIndex(key))
		}
		return result

	default:
		return v.Interface()
	}
}

// setFieldValue 设置字段值
func setFieldValue(fieldValue reflect.Value, mapValue interface{}) error {
	if mapValue == nil {
		return nil
	}

	mapValueReflect := reflect.ValueOf(mapValue)
	fieldType := fieldValue.Type()

	// 类型完全匹配，直接设置
	if mapValueReflect.Type() == fieldType {
		fieldValue.Set(mapValueReflect)
		return nil
	}

	// 处理指针类型
	if fieldType.Kind() == reflect.Ptr {
		if mapValueReflect.Type() == fieldType.Elem() {
			// 创建指针并设置值
			newPtr := reflect.New(fieldType.Elem())
			newPtr.Elem().Set(mapValueReflect)
			fieldValue.Set(newPtr)
			return nil
		}
		// 递归处理指针指向的类型
		if fieldValue.IsNil() {
			fieldValue.Set(reflect.New(fieldType.Elem()))
		}
		return setFieldValue(fieldValue.Elem(), mapValue)
	}

	// 处理结构体类型
	if fieldType.Kind() == reflect.Struct {
		// 特殊处理时间类型
		if fieldType == reflect.TypeOf(time.Time{}) {
			if timeVal, ok := mapValue.(time.Time); ok {
				fieldValue.Set(reflect.ValueOf(timeVal))
				return nil
			}
			if timeStr, ok := mapValue.(string); ok {
				if parsedTime, err := time.Parse(time.RFC3339, timeStr); err == nil {
					fieldValue.Set(reflect.ValueOf(parsedTime))
					return nil
				}
			}
		}

		// 处理嵌套结构体
		if mapData, ok := mapValue.(map[string]interface{}); ok {
			newStruct := reflect.New(fieldType)
			if err := MapToStruct(mapData, newStruct.Interface()); err != nil {
				return err
			}
			fieldValue.Set(newStruct.Elem())
			return nil
		}
	}

	// 处理切片类型
	if fieldType.Kind() == reflect.Slice {
		if sliceData, ok := mapValue.([]interface{}); ok {
			newSlice := reflect.MakeSlice(fieldType, len(sliceData), len(sliceData))
			for i, item := range sliceData {
				if err := setFieldValue(newSlice.Index(i), item); err != nil {
					return err
				}
			}
			fieldValue.Set(newSlice)
			return nil
		}
	}

	// 尝试类型转换
	if mapValueReflect.Type().ConvertibleTo(fieldType) {
		fieldValue.Set(mapValueReflect.Convert(fieldType))
		return nil
	}

	return fmt.Errorf("cannot convert %T to %s", mapValue, fieldType)
}
