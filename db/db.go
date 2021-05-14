package db

import (
	"fmt"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var err error

func Setup() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=%s",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Name,
		config.Database.Loc)

	var driver gorm.Dialector

	if config.Database.Dialect == "mysql" {
		driver = mysql.Open(dsn)
	} else if config.Database.Dialect == "postgres" {
		driver = postgres.Open(dsn)
	} else {
		logger.Fatal("Error db driver!")
	}

	newLogger := gormLogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		gormLogger.Config{
			SlowThreshold: time.Second,     // Slow SQL threshold
			LogLevel:      gormLogger.Info, // Log level
			Colorful:      true,            // Disable color
		},
	)

	db, err = gorm.Open(driver, &gorm.Config{
		SkipDefaultTransaction: false,
		PrepareStmt:            false,
		Logger:                 newLogger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   config.Database.TablePrefix,
			SingularTable: true, // use singular table name, table for `User` would be `user` with this option enabled
		},
	})
	if err != nil {
		logger.Error(err)
		logger.Fatal("Failed to connect to database!")
	}

	// Get generic database object sql.DB to use its functions
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error(err)
		logger.Fatal("Failed to connect to database 1!")
	}

	if sqlDB == nil {
		logger.Fatal("Failed to connect to database 2!")
		return
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(config.Database.MaxIdle)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(config.Database.MaxOpen)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)
}

func GetDB() *gorm.DB {
	return db
}

func CloseDB() {
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("Failed to close database!")
	}

	if sqlDB == nil {
		logger.Fatal("Failed to close database!")
		return
	}

	_ = sqlDB.Close()
}
