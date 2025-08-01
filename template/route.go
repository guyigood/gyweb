package main

import (
	"{project_name}/controller/dbcommon"
	"{project_name}/controller/sysbase"

	"github.com/guyigood/gyweb/core/engine"
)

func RegRoute(r *engine.Engine) {

	auth := r.Group("/api/auth")
	{
		auth.POST("/login", sysbase.Login)
		auth.POST("/logout", sysbase.Logout)
		auth.GET("/userinfo", sysbase.UserInfo)
		auth.GET("/getmenu", sysbase.GetRoleMenu)

	}
	// 数据库通用操作路由
	db := r.Group("/api/db")
	{
		db.GET("/page/:table", dbcommon.Page)
		db.GET("/list/:table", dbcommon.List)
		db.GET("/detail/:table", dbcommon.Detail)
		db.POST("/save/:table", dbcommon.Save)
		db.GET("/delete/:table", dbcommon.Delete)
		db.GET("/build", dbcommon.BuildTable)
		db.POST("/update/:table", dbcommon.UpdateData)
		db.POST("/batchupdate/:table", dbcommon.BatchUpdate)
		db.GET("/clearcache", dbcommon.ClearCache)

	}

}
