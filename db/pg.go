package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

var PG *gorm.DB

func InitPg() {
	log.Println("Connecting to Postgresql")
	var err error
	dsn := "host=127.0.0.1 user=root password=root dbname=swap port=8432 sslmode=disable"
	// 创建日志配置
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的地方，这里是标准输出）
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // Log level（Info级别会打印所有SQL）
			Colorful:      true,        // 彩色打印
		},
	)

	PG, err = gorm.Open(postgres.New(
		postgres.Config{
			DSN: dsn, // DSN data source name
		}), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}
	_ = PG.Callback().Row().After("gorm:row").Register("after_row", After)

	log.Println("Connected to Postgresql")
}

func After(db *gorm.DB) {
	db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
	sql := db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
	log.Println(sql)
}
