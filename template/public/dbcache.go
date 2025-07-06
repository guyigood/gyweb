package public

import (
	"errors"
	"fmt"
	"thermometer/model"

	"github.com/guyigood/gyweb/core/utils/datatype"
)

func GetTbInfo() {
	Db := GetDbConnection()
	defer Db.Close()
	data, err := Db.Table("sys_table_info").All()
	if err != nil {
		fmt.Println(err)
		return
	}
	Tbinfo = make([]model.GLobalTbInfo, len(data))
	for i, v := range data {
		Tbinfo[i].TableName = datatype.TypetoStr(v["table_name"])
		Tbinfo[i].ModuleName = datatype.TypetoStr(v["module_name"])
		Tbinfo[i].JoinTable = datatype.TypetoStr(v["join_tables"])
		Tbinfo[i].JoinField = datatype.TypetoStr(v["join_field_alias"])
		Tbinfo[i].PrimaryKey = datatype.TypetoStr(v["pk_field"])
		fData, err1 := Db.Table("sys_field_info").Where("table_id=?", v["id"]).All()
		if err1 != nil {
			fmt.Println(err1)
			return
		}
		Tbinfo[i].FdInfo = make([]model.GLobalFdInfo, len(fData))
		for j, v1 := range fData {
			Tbinfo[i].FdInfo[j].FieldName = datatype.TypetoStr(v1["field_name"])
			Tbinfo[i].FdInfo[j].FieldType = datatype.TypetoStr(v1["field_type"])
			Tbinfo[i].FdInfo[j].IsSearchable, _ = datatype.TypetoBool(v1["is_searchable"])
			Tbinfo[i].FdInfo[j].IsRequired, _ = datatype.TypetoBool(v1["is_required"])
			Tbinfo[i].FdInfo[j].IsUnique, _ = datatype.TypetoBool(v1["is_unique"])
			Tbinfo[i].FdInfo[j].IsActive, _ = datatype.TypetoBool(v1["is_active"])
			Tbinfo[i].FdInfo[j].QueryType = datatype.TypetoStr(v1["query_type"])
			Tbinfo[i].FdInfo[j].IsPk, _ = datatype.TypetoBool(v1["is_pk"])
		}
	}

}

func GetTbInfoByTableName(moduleName string) (model.GLobalTbInfo, error) {
	for _, v := range Tbinfo {
		if v.ModuleName == moduleName {
			return v, nil
		}
	}
	return model.GLobalTbInfo{}, errors.New("table not found")
}
