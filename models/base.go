package models

import (
	"database/sql"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

type DBInit struct {
	Db *sql.DB
}

func DBInitNew() *DBInit {
	db, _ := orm.GetDB()
	dbInit := DBInit{Db: db}
	return &dbInit
}

func (dbInit *DBInit) InitDatabase() {
	if dbInit.Db == nil {
		beego.LoadAppConfig("ini", "../conf/db.conf")
		orm.RegisterDriver("mysql", orm.DRMySQL)
		dbType := "mysql"
		host:=beego.AppConfig.String("mysql::host")
		port:=beego.AppConfig.String("mysql::port")
		user:=beego.AppConfig.String("mysql::user")
		pass:=beego.AppConfig.String("mysql::passw")
		dbName:=beego.AppConfig.String("mysql::dbName")
		dsn := user + ":" +pass +"@tcp("+host+":"+port+")/"+dbName+"?charset=utf8"
		maxIdle, _ := beego.AppConfig.Int("maxIdle")
		maxConn, _ := beego.AppConfig.Int("maxConn")
		logs.Info("Connect to [%v] database with conn: [%v] \n", dbType, dsn)
		orm.RegisterDataBase("default", dbType, dsn, maxIdle, maxConn)
		orm.Debug = false
	}
}