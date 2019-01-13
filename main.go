package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/kernelgarden/diet/router"
	"github.com/kernelgarden/goutils/config"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	// Read config file
	curPath, err := getCurPath()
	if err != nil {
		panic(err)
	}

	var c Config
	err = config.Read(curPath, "config", &c)
	if err != nil {
		panic(err)
	}

	// Init DB
	var dbURI string
	if c.Debug {
		dbURI = fmt.Sprintf("%s:%s@/%s_dev?charset=utf8", c.Database.Username, c.Database.Password, c.Database.Name)
	} else {
		dbURI = fmt.Sprintf("%s:%s@/%s?charset=utf8", c.Database.Username, c.Database.Password, c.Database.Name)
	}

	var dbType string
	if c.Database.Driver == "" {
		dbType = "mysql"
	} else {
		dbType = c.Database.Driver
	}

	db, err := initDB(dbType, dbURI)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	e := echo.New()

	router.InitRoutes(e)

	e.Static("/static", "static")
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	e.Debug = c.Debug

	var port string
	if c.Httpport == "" {
		port = "3030"
	} else {
		port = c.Httpport
	}

	if err := e.Start(fmt.Sprintf(":%s", port)); err != nil {
		log.Println(err)
	}
}

func initDB(driver, connection string) (*xorm.Engine, error) {
	db, err := xorm.NewEngine(driver, connection)
	if err != nil {
		return nil, err
	}

	if driver == "sqlite3" {
		runtime.GOMAXPROCS(1)
	}

	// TODO: sync db with model

	return db, nil
}

type Config struct {
	Database struct {
		Driver   string
		Username string
		Password string
		Name     string
		Logger   string
	}

	Behaviorlog struct {
		Kafka string
	}

	Debug    bool
	Service  string
	Httpport string
}

func getCurPath() (string, error) {
	curPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return curPath, nil
}
