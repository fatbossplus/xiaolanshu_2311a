// 工单编号: 微服务框架搭建
// 小蓝书互动社交平台 - MySQL数据库连接模块
// 使用GORM实现数据库操作，支持连接池配置
package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sync"
)

var (
	mysqlDB   *gorm.DB
	mysqlOnce sync.Once
)

// InitMySQL 初始化MySQL数据库连接
func InitMySQL() *gorm.DB {
	dsn := "root:123456@tcp(127.0.0.1:3306)/xiaolanshu_2311a?charset=utf8mb4&parseTime=True&loc=Local"
	mysqlOnce.Do(func() {
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			panic("数据库连接失败")
		}
		mysqlDB = db
	})

	return mysqlDB
}

// GetMySQL 获取MySQL数据库连接
func GetMySQL() *gorm.DB {
	if mysqlDB == nil {
		panic("MySQL未初始化，请先调用InitMySQL")
	}
	return mysqlDB
}
