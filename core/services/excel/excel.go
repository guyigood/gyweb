package excel

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExcelService Excel服务
type ExcelService struct {
	file *excelize.File
}

// ColumnMap 列映射
type ColumnMap struct {
	Name     string      `json:"name"`      // 列名
	Field    string      `json:"field"`     // 字段名
	Required bool        `json:"required"`  // 是否必填
	DataType string      `json:"data_type"` // 数据类型: string, int, float, bool, time
	Format   string      `json:"format"`    // 格式（时间格式等）
	Default  interface{} `json:"default"`   // 默认值
}

// ImportOptions 导入选项
type ImportOptions struct {
	SheetName    string                  `json:"sheet_name"`  // 工作表名称，默认为第一个
	StartRow     int                     `json:"start_row"`   // 开始行，默认为2（跳过标题行）
	HeaderRow    int                     `json:"header_row"`  // 标题行，默认为1
	MaxRows      int                     `json:"max_rows"`    // 最大行数，0表示无限制
	ColumnMaps   []ColumnMap             `json:"column_maps"` // 列映射
	ValidateFunc func(interface{}) error `json:"-"`           // 验证函数
}

// ExportOptions 导出选项
type ExportOptions struct {
	SheetName   string                        `json:"sheet_name"`   // 工作表名称
	Headers     []string                      `json:"headers"`      // 表头
	ColumnMaps  []ColumnMap                   `json:"column_maps"`  // 列映射
	StyleConfig *StyleConfig                  `json:"style_config"` // 样式配置
	FormatFunc  func(interface{}) interface{} `json:"-"`            // 格式化函数
}

// StyleConfig 样式配置
type StyleConfig struct {
	HeaderStyle  *CellStyle `json:"header_style"`  // 标题行样式
	DataStyle    *CellStyle `json:"data_style"`    // 数据行样式
	ColumnWidths []float64  `json:"column_widths"` // 列宽
}

// CellStyle 单元格样式
type CellStyle struct {
	Font      *FontStyle      `json:"font"`      // 字体样式
	Fill      *FillStyle      `json:"fill"`      // 填充样式
	Alignment *AlignmentStyle `json:"alignment"` // 对齐样式
	Border    *BorderStyle    `json:"border"`    // 边框样式
}

// FontStyle 字体样式
type FontStyle struct {
	Bold   bool   `json:"bold"`   // 粗体
	Italic bool   `json:"italic"` // 斜体
	Size   int    `json:"size"`   // 字体大小
	Color  string `json:"color"`  // 字体颜色
	Family string `json:"family"` // 字体族
}

// FillStyle 填充样式
type FillStyle struct {
	Type    string `json:"type"`    // 填充类型
	Pattern int    `json:"pattern"` // 图案
	Color   string `json:"color"`   // 颜色
}

// AlignmentStyle 对齐样式
type AlignmentStyle struct {
	Horizontal string `json:"horizontal"` // 水平对齐
	Vertical   string `json:"vertical"`   // 垂直对齐
	WrapText   bool   `json:"wrap_text"`  // 文本换行
}

// BorderStyle 边框样式
type BorderStyle struct {
	Type  string `json:"type"`  // 边框类型
	Color string `json:"color"` // 边框颜色
}

// ValidationError 验证错误
type ValidationError struct {
	Row     int    `json:"row"`     // 错误行号
	Column  string `json:"column"`  // 错误列名
	Message string `json:"message"` // 错误信息
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("第%d行，列%s：%s", e.Row, e.Column, e.Message)
}

// ImportResult 导入结果
type ImportResult struct {
	Data         []interface{}     `json:"data"`          // 导入的数据
	SuccessCount int               `json:"success_count"` // 成功数量
	ErrorCount   int               `json:"error_count"`   // 错误数量
	Errors       []ValidationError `json:"errors"`        // 错误列表
}

// NewExcelService 创建Excel服务
func NewExcelService() *ExcelService {
	return &ExcelService{
		file: excelize.NewFile(),
	}
}

// NewExcelServiceWithFile 从文件创建Excel服务
func NewExcelServiceWithFile(filename string) (*ExcelService, error) {
	file, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, err
	}
	return &ExcelService{file: file}, nil
}

// NewExcelServiceWithReader 从Reader创建Excel服务
func NewExcelServiceWithReader(data []byte) (*ExcelService, error) {
	file, err := excelize.OpenReader(strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	return &ExcelService{file: file}, nil
}

// ImportData 导入数据
func (e *ExcelService) ImportData(options *ImportOptions, target interface{}) (*ImportResult, error) {
	if options == nil {
		options = &ImportOptions{
			StartRow:  2,
			HeaderRow: 1,
		}
	}

	// 检查target是否为slice的指针
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Slice {
		return nil, errors.New("target必须是slice的指针")
	}

	sliceValue := targetValue.Elem()
	elementType := sliceValue.Type().Elem()

	// 获取工作表名称
	sheetName := options.SheetName
	if sheetName == "" {
		sheetName = e.file.GetSheetName(0)
	}

	// 获取数据行
	rows, err := e.file.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return &ImportResult{Data: []interface{}{}, SuccessCount: 0, ErrorCount: 0}, nil
	}

	// 获取表头
	var headers []string
	if options.HeaderRow > 0 && len(rows) >= options.HeaderRow {
		headers = rows[options.HeaderRow-1]
	}

	// 建立列映射
	columnIndexMap := make(map[string]int)
	for i, header := range headers {
		columnIndexMap[strings.TrimSpace(header)] = i
	}

	result := &ImportResult{
		Data:   []interface{}{},
		Errors: []ValidationError{},
	}

	// 处理数据行
	startRow := options.StartRow
	if startRow <= 0 {
		startRow = 2
	}

	maxRows := options.MaxRows
	if maxRows <= 0 {
		maxRows = len(rows)
	}

	for i := startRow - 1; i < len(rows) && i < startRow-1+maxRows; i++ {
		row := rows[i]
		rowNum := i + 1

		// 创建目标对象
		element := reflect.New(elementType).Interface()

		if err := e.mapRowToStruct(row, element, options.ColumnMaps, columnIndexMap, rowNum, result); err != nil {
			result.ErrorCount++
			continue
		}

		// 自定义验证
		if options.ValidateFunc != nil {
			if err := options.ValidateFunc(element); err != nil {
				result.Errors = append(result.Errors, ValidationError{
					Row:     rowNum,
					Column:  "数据验证",
					Message: err.Error(),
				})
				result.ErrorCount++
				continue
			}
		}

		result.Data = append(result.Data, element)
		result.SuccessCount++
	}

	// 更新target
	for _, data := range result.Data {
		sliceValue = reflect.Append(sliceValue, reflect.ValueOf(data).Elem())
	}
	targetValue.Elem().Set(sliceValue)

	return result, nil
}

// ExportData 导出数据
func (e *ExcelService) ExportData(data interface{}, options *ExportOptions) error {
	if options == nil {
		return errors.New("导出选项不能为空")
	}

	// 检查data是否为slice
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() != reflect.Slice {
		return errors.New("data必须是slice类型")
	}

	sheetName := options.SheetName
	if sheetName == "" {
		sheetName = "Sheet1"
	}

	// 创建工作表
	sheetIndex, err := e.file.GetSheetIndex(sheetName)
	if err != nil {
		return err
	}
	if sheetIndex == -1 {
		_, err := e.file.NewSheet(sheetName)
		if err != nil {
			return err
		}
	}

	// 写入表头
	if len(options.Headers) > 0 {
		for i, header := range options.Headers {
			cell := e.getColumnName(i+1) + "1"
			e.file.SetCellValue(sheetName, cell, header)
		}
	}

	// 设置样式
	if options.StyleConfig != nil {
		e.applyStyles(sheetName, options.StyleConfig, len(options.Headers), dataValue.Len())
	}

	// 写入数据
	for i := 0; i < dataValue.Len(); i++ {
		rowData := dataValue.Index(i).Interface()
		rowNum := i + 2 // 从第2行开始（第1行是表头）

		if options.FormatFunc != nil {
			rowData = options.FormatFunc(rowData)
		}

		if err := e.writeRowData(sheetName, rowNum, rowData, options.ColumnMaps); err != nil {
			return err
		}
	}

	return nil
}

// SaveToFile 保存到文件
func (e *ExcelService) SaveToFile(filename string) error {
	return e.file.SaveAs(filename)
}

// GetBytes 获取文件字节数据
func (e *ExcelService) GetBytes() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	if err := e.file.Write(buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Close 关闭文件
func (e *ExcelService) Close() error {
	return e.file.Close()
}

// mapRowToStruct 将行数据映射到结构体
func (e *ExcelService) mapRowToStruct(row []string, target interface{}, columnMaps []ColumnMap, columnIndexMap map[string]int, rowNum int, result *ImportResult) error {
	targetValue := reflect.ValueOf(target).Elem()
	targetType := targetValue.Type()

	for _, columnMap := range columnMaps {
		columnIndex, exists := columnIndexMap[columnMap.Name]
		if !exists {
			if columnMap.Required {
				result.Errors = append(result.Errors, ValidationError{
					Row:     rowNum,
					Column:  columnMap.Name,
					Message: "列不存在",
				})
				return errors.New("必填列不存在")
			}
			continue
		}

		var cellValue string
		if columnIndex < len(row) {
			cellValue = strings.TrimSpace(row[columnIndex])
		}

		// 检查必填字段
		if columnMap.Required && cellValue == "" {
			result.Errors = append(result.Errors, ValidationError{
				Row:     rowNum,
				Column:  columnMap.Name,
				Message: "必填字段不能为空",
			})
			return errors.New("必填字段为空")
		}

		// 使用默认值
		if cellValue == "" && columnMap.Default != nil {
			cellValue = fmt.Sprintf("%v", columnMap.Default)
		}

		// 查找字段
		field := targetValue.FieldByName(columnMap.Field)
		if !field.IsValid() {
			// 尝试通过tag查找字段
			for i := 0; i < targetType.NumField(); i++ {
				structField := targetType.Field(i)
				if tag := structField.Tag.Get("excel"); tag == columnMap.Field {
					field = targetValue.Field(i)
					break
				}
				if tag := structField.Tag.Get("json"); tag == columnMap.Field {
					field = targetValue.Field(i)
					break
				}
			}
		}

		if !field.IsValid() || !field.CanSet() {
			continue
		}

		// 类型转换
		if err := e.setFieldValue(field, cellValue, columnMap, rowNum, result); err != nil {
			return err
		}
	}

	return nil
}

// setFieldValue 设置字段值
func (e *ExcelService) setFieldValue(field reflect.Value, value string, columnMap ColumnMap, rowNum int, result *ImportResult) error {
	if value == "" {
		return nil
	}

	dataType := columnMap.DataType
	if dataType == "" {
		// 根据字段类型推断
		switch field.Kind() {
		case reflect.String:
			dataType = "string"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			dataType = "int"
		case reflect.Float32, reflect.Float64:
			dataType = "float"
		case reflect.Bool:
			dataType = "bool"
		default:
			if field.Type() == reflect.TypeOf(time.Time{}) {
				dataType = "time"
			} else {
				dataType = "string"
			}
		}
	}

	switch dataType {
	case "string":
		field.SetString(value)

	case "int":
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Row:     rowNum,
				Column:  columnMap.Name,
				Message: fmt.Sprintf("无法转换为整数：%s", value),
			})
			return err
		}
		field.SetInt(intValue)

	case "float":
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Row:     rowNum,
				Column:  columnMap.Name,
				Message: fmt.Sprintf("无法转换为浮点数：%s", value),
			})
			return err
		}
		field.SetFloat(floatValue)

	case "bool":
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			// 尝试其他常见的布尔值表示
			lowerValue := strings.ToLower(value)
			switch lowerValue {
			case "是", "yes", "y", "1":
				boolValue = true
			case "否", "no", "n", "0":
				boolValue = false
			default:
				result.Errors = append(result.Errors, ValidationError{
					Row:     rowNum,
					Column:  columnMap.Name,
					Message: fmt.Sprintf("无法转换为布尔值：%s", value),
				})
				return err
			}
		}
		field.SetBool(boolValue)

	case "time":
		format := columnMap.Format
		if format == "" {
			format = "2006-01-02 15:04:05"
		}

		timeValue, err := time.Parse(format, value)
		if err != nil {
			// 尝试其他常见格式
			formats := []string{
				"2006-01-02",
				"2006/01/02",
				"01/02/2006",
				"2006-01-02 15:04:05",
				"2006/01/02 15:04:05",
			}

			for _, f := range formats {
				if timeValue, err = time.Parse(f, value); err == nil {
					break
				}
			}

			if err != nil {
				result.Errors = append(result.Errors, ValidationError{
					Row:     rowNum,
					Column:  columnMap.Name,
					Message: fmt.Sprintf("无法转换为时间：%s", value),
				})
				return err
			}
		}
		field.Set(reflect.ValueOf(timeValue))

	default:
		field.SetString(value)
	}

	return nil
}

// writeRowData 写入行数据
func (e *ExcelService) writeRowData(sheetName string, rowNum int, data interface{}, columnMaps []ColumnMap) error {
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() == reflect.Ptr {
		dataValue = dataValue.Elem()
	}

	for i, columnMap := range columnMaps {
		columnName := e.getColumnName(i + 1)
		cell := columnName + strconv.Itoa(rowNum)

		var value interface{}

		// 查找字段值
		fieldValue := dataValue.FieldByName(columnMap.Field)
		if !fieldValue.IsValid() {
			// 尝试通过tag查找字段
			dataType := dataValue.Type()
			for j := 0; j < dataType.NumField(); j++ {
				structField := dataType.Field(j)
				if tag := structField.Tag.Get("excel"); tag == columnMap.Field {
					fieldValue = dataValue.Field(j)
					break
				}
				if tag := structField.Tag.Get("json"); tag == columnMap.Field {
					fieldValue = dataValue.Field(j)
					break
				}
			}
		}

		if fieldValue.IsValid() {
			value = fieldValue.Interface()

			// 格式化时间
			if columnMap.DataType == "time" || fieldValue.Type() == reflect.TypeOf(time.Time{}) {
				if timeValue, ok := value.(time.Time); ok {
					format := columnMap.Format
					if format == "" {
						format = "2006-01-02 15:04:05"
					}
					value = timeValue.Format(format)
				}
			}
		}

		e.file.SetCellValue(sheetName, cell, value)
	}

	return nil
}

// applyStyles 应用样式
func (e *ExcelService) applyStyles(sheetName string, styleConfig *StyleConfig, columnCount, rowCount int) {
	// 设置列宽
	if len(styleConfig.ColumnWidths) > 0 {
		for i, width := range styleConfig.ColumnWidths {
			if i < columnCount {
				columnName := e.getColumnName(i + 1)
				e.file.SetColWidth(sheetName, columnName, columnName, width)
			}
		}
	}

	// 应用表头样式
	if styleConfig.HeaderStyle != nil {
		headerStyleID := e.createCellStyle(styleConfig.HeaderStyle)
		if headerStyleID != 0 {
			for i := 1; i <= columnCount; i++ {
				cell := e.getColumnName(i) + "1"
				e.file.SetCellStyle(sheetName, cell, cell, headerStyleID)
			}
		}
	}

	// 应用数据样式
	if styleConfig.DataStyle != nil {
		dataStyleID := e.createCellStyle(styleConfig.DataStyle)
		if dataStyleID != 0 {
			for row := 2; row <= rowCount+1; row++ {
				for col := 1; col <= columnCount; col++ {
					cell := e.getColumnName(col) + strconv.Itoa(row)
					e.file.SetCellStyle(sheetName, cell, cell, dataStyleID)
				}
			}
		}
	}
}

// createCellStyle 创建单元格样式
func (e *ExcelService) createCellStyle(cellStyle *CellStyle) int {
	style := &excelize.Style{}

	if cellStyle.Font != nil {
		style.Font = &excelize.Font{
			Bold:   cellStyle.Font.Bold,
			Italic: cellStyle.Font.Italic,
			Size:   float64(cellStyle.Font.Size),
			Color:  cellStyle.Font.Color,
			Family: cellStyle.Font.Family,
		}
	}

	if cellStyle.Fill != nil {
		style.Fill = excelize.Fill{
			Type:    cellStyle.Fill.Type,
			Pattern: cellStyle.Fill.Pattern,
			Color:   []string{cellStyle.Fill.Color},
		}
	}

	if cellStyle.Alignment != nil {
		style.Alignment = &excelize.Alignment{
			Horizontal: cellStyle.Alignment.Horizontal,
			Vertical:   cellStyle.Alignment.Vertical,
			WrapText:   cellStyle.Alignment.WrapText,
		}
	}

	if cellStyle.Border != nil {
		style.Border = []excelize.Border{
			{Type: "left", Color: cellStyle.Border.Color, Style: 1},
			{Type: "top", Color: cellStyle.Border.Color, Style: 1},
			{Type: "bottom", Color: cellStyle.Border.Color, Style: 1},
			{Type: "right", Color: cellStyle.Border.Color, Style: 1},
		}
	}

	styleID, _ := e.file.NewStyle(style)
	return styleID
}

// getColumnName 获取列名（A, B, C, ..., AA, AB, ...）
func (e *ExcelService) getColumnName(column int) string {
	name := ""
	for column > 0 {
		column--
		name = string(rune('A'+(column%26))) + name
		column /= 26
	}
	return name
}

// GetSheetNames 获取所有工作表名称
func (e *ExcelService) GetSheetNames() []string {
	return e.file.GetSheetList()
}

// SetActiveSheet 设置活动工作表
func (e *ExcelService) SetActiveSheet(name string) error {
	index, err := e.file.GetSheetIndex(name)
	if err != nil {
		return err
	}
	if index == -1 {
		return fmt.Errorf("工作表 %s 不存在", name)
	}
	e.file.SetActiveSheet(index)
	return nil
}

// AddSheet 添加工作表
func (e *ExcelService) AddSheet(name string) error {
	_, err := e.file.NewSheet(name)
	return err
}

// DeleteSheet 删除工作表
func (e *ExcelService) DeleteSheet(name string) error {
	return e.file.DeleteSheet(name)
}

// GetCellValue 获取单元格值
func (e *ExcelService) GetCellValue(sheetName, cell string) (string, error) {
	return e.file.GetCellValue(sheetName, cell)
}

// SetCellValue 设置单元格值
func (e *ExcelService) SetCellValue(sheetName, cell string, value interface{}) error {
	return e.file.SetCellValue(sheetName, cell, value)
}

// GetRows 获取所有行数据
func (e *ExcelService) GetRows(sheetName string) ([][]string, error) {
	return e.file.GetRows(sheetName)
}

// GetCols 获取所有列数据
func (e *ExcelService) GetCols(sheetName string) ([][]string, error) {
	return e.file.GetCols(sheetName)
}
