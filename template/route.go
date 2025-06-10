package main

import (
	"{firstweb}/controller/dbcommon"
	"{firstweb}/controller/sysbase"

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
		db.POST("/page", dbcommon.Page)
		db.POST("/list", dbcommon.List)
		db.GET("/detail", dbcommon.Detail)
		db.POST("/save", dbcommon.Save)
		db.POST("/delete", dbcommon.Delete)
	}
}
