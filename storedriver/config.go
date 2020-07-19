package storedriver

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)
type Db struct {
	Rediscon  RedisClient `yaml:"redis"`
	Mysqlcon  MysqlConfig  `yaml:"mysql"`
	LocatDir  string  `yaml:"locatdir"`
	Metric  []Metric `yaml:"maps"`
}
type MonitorConfig struct {
	Recyle int  `yaml:"recyle"`
	GroupNum int  `yaml:"groupnum"`
	CpuNum   int `yaml:"cpunum"`
	Db   Db     `yaml:"db"`
}

type Metric struct {
	Metric string `yaml:"metric"`
	Redis  string `yaml:"redis"`
	Mysql  string `yaml:"mysql"`
	Cate   string  `yaml:"type"`
}

func ParseConfig(path string) (*MonitorConfig,error){
	var config MonitorConfig
	f, er:=os.Open(path)
	defer f.Close()
	if er != nil {
		return nil, er
	}
	data,_:=ioutil.ReadAll(f)
	err:=yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config,nil

}