package config

import (
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xormplus/xorm"
	"time"
)

type Config struct {
	DB struct {
		Host    string
		User    string
		Pwd     string
		Db      string
		Show    bool
		Port    int
		MaxOpen int
		MaxIdle int
	}
	Redis struct {
		Dns      string
		MinIdle  int
		PoolSize int
	}

	Mgo struct {
		Dns string
	}
}

var (
	EngMgo *mgo.Session
	EngDb  *xorm.Engine
	EngRds *redis.Client
)

func InitConfig(path string) {
	config := Config{}
	Load(path, &config)
	config.loadDb()
	config.loadRedis()
	//config.loadMgo()
}

func (c *Config) loadDb() {
	fmt.Println(c.DB)
	var err error
	dns := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		c.DB.User,
		c.DB.Pwd,
		c.DB.Host,
		c.DB.Db)

	EngDb, err = xorm.NewEngine("mysql", dns)

	ping := EngDb.Ping()
	if ping != nil || err != nil {
		panic(ping)
	}
	EngDb.SetMaxIdleConns(c.DB.MaxIdle)
	EngDb.SetMaxOpenConns(c.DB.MaxOpen)
	EngDb.ShowSQL(c.DB.Show)

}

func (c Config) loadRedis() {

	EngRds = redis.NewClient(&redis.Options{
		Addr:         c.Redis.Dns,
		Password:     "", // no password set
		DB:           0,  // use default DB
		PoolSize:     c.Redis.PoolSize,
		MinIdleConns: c.Redis.MinIdle,
	})

	_, err := EngRds.Ping().Result()

	//a := EngRds.Info()
	//demo(fmt.Sprintf("%s", a))

	if err != nil {
		panic(err)
	}

}

func (c *Config) loadMgo() {
	var err error
	dialInfo := &mgo.DialInfo{
		Addrs:     []string{c.Mgo.Dns},
		Source:    "mdata",
		Username:  "xiaohan",
		Password:  "xiaohanmongodata",
		Timeout:   60 * time.Second,
		PoolLimit: 100,
	}

	EngMgo, err = mgo.DialWithInfo(dialInfo)

	//
	if err != nil {
		panic(err)
	}
}
