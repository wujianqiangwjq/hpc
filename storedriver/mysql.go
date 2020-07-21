package storedriver

import (
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gohouse/gorose"
)

type MysqlConfig struct {
	Driver       string `yaml:"driver"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Database     string `yaml:"dbname"`
	Address      string `yaml:"address"`
	Port         string `yaml:"port"`
	IdleConns    int    `yaml:"idleconns"`
	MaxOpenConns int    `yaml:"maxopenconns"`
	Table        string `yaml:"table"`
}

func Open(config *MysqlConfig) (*gorose.Engin, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true",
		config.Username, config.Password, config.Address, config.Port, config.Database)
	con := &gorose.Config{
		Driver:          config.Driver,
		Dsn:             dsn,
		SetMaxOpenConns: config.IdleConns,
		SetMaxIdleConns: config.MaxOpenConns,
	}
	return gorose.Open(con)
}

func InsertData(engin *gorose.Engin, table string, data map[string]interface{}) error {
	orm := engin.NewOrm()
	var jobid int
	var err error = errors.New("jobid can't find")
	jobidinterface, ok := data["jobid"]
	if !ok {
		return err
	}
	switch jobidinterface.(type) {
	case float64:
		jobid = int(jobidinterface.(float64))
	case int:
		jobid = jobidinterface.(int)
	}
	count, _ := orm.Table(table).Where("jobid", "=", jobid).Count()
	if count == 0 {
		_, err := orm.Table(table).Data(data).Insert()
		return err
	}
	return nil

}
