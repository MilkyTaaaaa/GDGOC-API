package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDatabase() {
	var err error
	
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	gormLogger := logger.Default.LogMode(logger.Info)
	if os.Getenv("APP_ENV") == "production" {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger,
			NowFunc: func() time.Time {
				loc, _ := time.LoadLocation("Asia/Jakarta")
				return time.Now().In(loc)
			},
		})

		if err == nil {
			break
		}

		log.Printf("Gagal koneksi database (%d/%d): %v", i+1, maxRetries, err)
		time.Sleep(time.Second * 2)
	}

	if err != nil {
		log.Fatal("Gagal koneksi :", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Gagal mendapat instance database:", err)
	}

	sqlDB.SetMaxIdleConns(10)          
	sqlDB.SetMaxOpenConns(100)         
	sqlDB.SetConnMaxLifetime(time.Hour) 

	// Tes koneksi
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Gagal ping database:", err)
	}

	log.Println("Koneksyen Berhasil!!!")
}

func GetDB() *gorm.DB {
	return DB
}

func CloseDatabase() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Printf("Gagal mendapat instance database: %v", err)
			return
		}
		
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error saat menutup database: %v", err)
			return
		}
		
		fmt.Println("Koneksyen database ditutup")
	}
}