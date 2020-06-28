package main

import (
	"fmt"
	"influx"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"
	iclient "github.com/influxdata/influxdb1-client/v2"
	"github.com/mxj"
)

type Metric struct {
	hostname    string
	load        float64
	bytes_in    float64
	bytes_out   float64
	cpu_num     int
	mem_total   float64
	mem_buffers float64
	mem_cached  float64
	mem_free    float64
	disk_total  float64
	disk_free   float64
	cpu_idle    float64
	tn          int
	tmax        int
}

func (m *Metric) Println() {
	fmt.Println("hostname:", m.hostname,
		" load:", m.load, " bytes_in:", m.bytes_in, " bytes_out:", m.bytes_out,
		" cpu_num:", m.cpu_num, " mem_total:", m.mem_total, " mem_buffers:", m.mem_buffers,
		" mem_cached:", m.mem_cached, " mem_free:", m.mem_free, " disk_total:", m.disk_total,
		" disk_free:", m.disk_free, " cpu_idle:", m.cpu_idle)
}
func (m *Metric) GetCpuUsage() float64 {
	data := 100 - m.cpu_idle
	ret, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", data), 64)
	return ret
}
func (m *Metric) GetMemUsed() float64 {
	data := m.mem_total - m.mem_free - m.mem_buffers - m.mem_cached
	ret, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", data), 64)
	return ret
}
func (m *Metric) GetMemUsage() float64 {
	data := m.GetDiskUsed() / m.mem_total
	ret, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", data), 64)
	return ret
}
func (m *Metric) GetDiskUsed() float64 {
	data := m.disk_total - m.disk_free
	ret, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", data), 64)
	return ret
}
func (m *Metric) GetDiskUsage() float64 {
	data := m.GetDiskUsed() / m.disk_total
	ret, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", data), 64)
	return ret
}
func (m *Metric) IsActive() bool {
	return m.tn < 4*m.tmax
}
var monitorconfig MonitorConfig

func init() {
	f,_ := os.Open("c.yaml")
	defer f.Close()
	data,eror:= ioutil.ReadAll(f)
	if eror != nil {
		panic(eror)
	}
	er:=yaml.Unmarshal(data,&monitorconfig)
	if er != nil {
		panic(er)
	}

}

func one_rycle(client iclient.Client, config MonitorHost, influx InfluxConfig) {
	batchpoints, _ := influx.GetBatchPoints(config.Rp)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", config.Host)
	if err != nil {
		log.Fatal(err)
		return
   }
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()
	result, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Fatal(err)
		return
	}
	mxj.XmlCharsetReader = charset.NewReader
	mv, err := mxj.NewMapXml(result)
	if err != nil {
		log.Fatal(err)
		return
	}
	hosts, err := mv.ValuesForPath("GANGLIA_XML.CLUSTER.HOST")
	if err != nil {
		log.Fatal(err)
		return
	}
	for _, v := range hosts {
		host_metric := new(Metric)
		host_data := v.(map[string]interface{})
		host_metric.hostname = host_data["-NAME"].(string)
		tn_s := host_data["-TN"].(string)
		tn, _ := strconv.Atoi(tn_s)
		host_metric.tn = tn
		tmax_s := host_data["-TMAX"].(string)
		tmax, _ := strconv.Atoi(tmax_s)
		host_metric.tmax = tmax
		metrics := host_data["METRIC"].([]interface{})
		for _, item := range metrics {
			metric := item.(map[string]interface{})
			name := metric["-NAME"].(string)
			switch name {
			case "mem_cached":
				mem_cached := metric["-VAL"].(string)
				value, _ := strconv.ParseFloat(mem_cached, 2)
				host_metric.mem_cached = value
			case "bytes_in":
				bytes_in := metric["-VAL"].(string)
				value, _ := strconv.ParseFloat(bytes_in, 2)
				host_metric.bytes_in = value
			case "disk_free":
				disk_free := metric["-VAL"].(string)
				value, _ := strconv.ParseFloat(disk_free, 2)
				host_metric.disk_free = value
			case "disk_total":
				disk_total := metric["-VAL"].(string)
				value, _ := strconv.ParseFloat(disk_total, 2)
				host_metric.disk_total = value
			case "mem_total":
				mem_total := metric["-VAL"].(string)
				value, _ := strconv.ParseFloat(mem_total, 2)
				host_metric.mem_total = value
			case "bytes_out":
				bytes_out := metric["-VAL"].(string)
				value, _ := strconv.ParseFloat(bytes_out, 2)
				host_metric.bytes_out = value
			case "mem_free":
				mem_free := metric["-VAL"].(string)
				value, _ := strconv.ParseFloat(mem_free, 2)
				host_metric.mem_free = value
			case "cpu_idle":
				cpu_idle := metric["-VAL"].(string)
				value, _ := strconv.ParseFloat(cpu_idle, 2)
				host_metric.cpu_idle = value
			case "mem_buffers":
				mem_buffers := metric["-VAL"].(string)
				value, _ := strconv.ParseFloat(mem_buffers, 2)
				host_metric.mem_buffers = value
			case "cpu_num":
				cpu_num := metric["-VAL"].(string)
				value, _ := strconv.Atoi(cpu_num)
				host_metric.cpu_num = value
			case "load_one":
				load := metric["-VAL"].(string)
				value, _ := strconv.ParseFloat(load, 2)
				host_metric.load = value
			}
		}
		if host_metric.IsActive() {
			tags := map[string]string{
				"host": host_metric.hostname,
			}
			fields := map[string]interface{}{
				"load":       host_metric.load,
				"cpu_usage":  host_metric.GetCpuUsage(),
				"cpu_num":    host_metric.cpu_num,
				"mem_usage":  host_metric.GetMemUsage(),
				"mem_total":  host_metric.mem_total,
				"disk_total": host_metric.disk_total,
				"disk_usage": host_metric.GetDiskUsage(),
				"bytes_in":   host_metric.bytes_in,
				"bytes_out":  host_metric.bytes_out,
			}
			current_time := time.Now()
			point, _ := influx.AddPoints(config.Keys, tags, fields, current_time)
			batchpoints.AddPoint(point)
		}

	}
	client.Write(batchpoints)
}
func main() {
	client, _ := monitorconfig.Influxconfig.Connect(false)
	//one_rycle(client, influxconfig)

}
