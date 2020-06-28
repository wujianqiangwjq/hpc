package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	_ "github.com/influxdata/influxdb1-client"
	client "github.com/influxdata/influxdb1-client/v2"
	"time"
)

type InfluxConfig struct{
	Username string  `yaml:"username"`
	Address  string  `yaml:"address"`
	Password   string `yaml:"password"`
	Database   string `yaml:"database"`
	Timeout  int   `yaml:"timeout"`
}
func (ic InfluxConfig) Connect(once bool) (client.Client, error) {
	timeout := time.Duration(ic.Timeout)
	for {
		influxclient, err := client.NewHTTPClient(client.HTTPConfig{
			Addr:     ic.Address,
			Username: ic.Username,
			Password: ic.Password,
			Timeout: timeout ,
		})
		if err != nil {
			log.Fatal(err)
		} else {
			return influxclient, nil
		}
		if once {
			return nil, err
		}

	}
}

func (ic InfluxConfig) GetBatchPoints(rp string) (client.BatchPoints, error) {
	batchpoints, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:        ic.Database,
		RetentionPolicy: rp,
	})
	return batchpoints, err
}

type  Values struct {
	Re   bool  `yaml:"re"`
	Name string  `yaml:"name"`
	Section string `yaml:"section"`
}

type KeyValue struct {
	Key  string `yaml:"key"`
	Value Values `yaml:"value"`
}

type  MetricConfig struct {
	Name string `yaml:"name"`

	Tags  []KeyValue `yaml:"tags"`
	Fields []KeyValue `yaml:"fields"`
}
type MonitorHost struct {
  Host  string `yaml:"host"`
  Rp  string `yaml:"rp"`
  Metrics []MetricConfig `yaml:"metrics"`
}

type  MonitorConfig struct {
	Influxconfig  InfluxConfig `yaml:"influx"`
	Monitorhosts []MonitorHost `yaml:"hosts"`
	Recyletime   int  `yaml:"recyletime"`
}

var ComplieError = errors.New("regex error")
var KeyNotExist = errors.New("key not exist")
var SectionError = errors.New("Section is error")
func (v Values)GetReValue(data string) (string,error) {
	if v.Re {
		compile,err := regexp.Compile(v.Name)
		if err != nil {
			panic(err)
		}
		subdata:=compile.FindAllStringSubmatch(data,1)
		if len(subdata) > 0 {
			if len(subdata[0]) == 2{
				return subdata[0][1],nil
			}
		}

	}
	return "",ComplieError
}
func (v Values)GetValue(data map[string]interface{}) (string,error) {
	if v.Section == "host"{
		value,ok := data[v.Name]
		if !ok {
			return "",KeyNotExist
		}
		return value.(string),nil
	}
	if v.Section == "metric" {
		key := "-VAL"
		value,ok := data[key]
		if !ok {
			return "",KeyNotExist
		}
		return value.(string),nil
	}
	return "",SectionError
}

func main()  {
	var out MonitorConfig
	f,_ := os.Open("c.yaml")
	data,eror:= ioutil.ReadAll(f)
	if eror != nil {
		panic(eror)
	}
	er:=yaml.Unmarshal(data,&out)

fmt.Println(er)
fmt.Println(out)

}