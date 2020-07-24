package main

/*
#cgo LDFLAGS: -lslurm
#include "slurmclient.h"
*/
import "C"
import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gohouse/gorose"
	"github.com/storedriver"
)

var JobState = map[int]string{
	0:  "PENDING",
	1:  "RUNNING",
	2:  "SUSPENDED",
	3:  "COMPLETE",
	4:  "CANCELLED",
	5:  "FAILED",
	6:  "TIMEOUT",
	7:  "NODE_FAIL",
	8:  "PREEMPTED",
	9:  "BOOT_FAIL",
	10: "DEADLINE",
	11: "OOM",
	12: "END",
}

type Job struct {
	Jobid      uint32
	Jobname    string
	Queue      string
	Qtime      int64
	Starttime  int64
	Endtime    int64
	Userid     uint32
	Username   string
	Jobstatus  int
	Jobnodes   int
	Nodes      string
	Jobcpus    int
	Jobgpus    int
	Jobtres    string
	Jobmemory  uint64
	Joberrfile string
	Joboutfile string
	Jobcommand string
	Jobworkdir string
}

func GetGpus(tres string) int {
	gpus := strings.Split(tres, ":")
	if len(gpus) == 2 && gpus[0] == "gpu" {
		num, er := strconv.Atoi(gpus[1])
		if er == nil {
			return num
		}
	}
	return 0
}
func (job *Job) Set(pjob C.struct_privitejob) {
	job.Jobid = uint32(pjob.jobid)
	job.Jobname = C.GoString(pjob.name)
	job.Queue = C.GoString(pjob.partition)
	job.Qtime = int64(pjob.submit_time)
	job.Starttime = int64(pjob.start_time)
	job.Endtime = int64(pjob.end_time)
	job.Userid = uint32(pjob.user_id)
	jobuser, _ := user.LookupId(strconv.Itoa(int(job.Userid)))
	job.Username = jobuser.Username
	job.Jobstatus = int(pjob.job_state)
	job.Jobnodes = int(pjob.num_nodes)
	job.Jobcpus = int(pjob.num_cpus)
	job.Jobgpus = int(pjob.gpus)
	job.Jobtres = C.GoString(pjob.tres_per_node)
	job.Nodes = C.GoString(pjob.nodes)
	job.Jobmemory = uint64(pjob.memory_used)
	job.Jobworkdir = C.GoString(pjob.work_dir)
	job.Joberrfile = C.GoString(pjob.std_err)
	if job.Joberrfile == "" {
		job.Joberrfile = fmt.Sprintf("%s/slurm-%d.out", job.Jobworkdir, job.Jobid)
	}
	job.Joboutfile = C.GoString(pjob.std_out)
	if job.Joboutfile == "" {
		job.Joboutfile = fmt.Sprintf("%s/slurm-%d.out", job.Jobworkdir, job.Jobid)
	}
	job.Jobcommand = C.GoString(pjob.command)
	if job.Jobgpus == 0 && job.Jobtres != "" {
		job.Jobgpus = job.Jobnodes * GetGpus(job.Jobtres)
	}
}
func (job *Job) Write(table string) {
	state, ok := JobState[job.Jobstatus]
	if ok {
		data := map[string]interface{}{
			"jobid":      int(job.Jobid),
			"jobname":    job.Jobname,
			"queue":      job.Queue,
			"qtime":      job.Qtime,
			"starttime":  job.Starttime,
			"endtime":    job.Endtime,
			"submiter":   job.Username,
			"jobstatus":  state,
			"nodescount": job.Jobnodes,
			"cpuscount":  job.Jobcpus,
			"tres_text":  job.Jobtres,
			"memory":     job.Jobmemory,
			"gpus":       job.Jobgpus,
			"nodes":      job.Nodes,
			"errput":     job.Joberrfile,
			"output":     job.Joboutfile,
			"command":    job.Jobcommand,
		}
		if job.Jobstatus >= 3 {
			redisc, er := RedisPool.Acquire()
			if er == nil {
				mem, memer := storedriver.GetKey(redisc, strconv.Itoa(int(job.Jobid)), []string{"memory"})
				if memer == nil && len(mem) == 1 {
					if mem[0] != nil {
						memstr := mem[0].(string)
						data["memory"], _ = strconv.ParseInt(memstr, 10, 64)
					}
				}
			}

			inserterr := storedriver.InsertData(MysqlPool, table, data)
			if inserterr != nil {
				Logger.Println(inserterr)
				filerr := FilesHandle.Add(strconv.Itoa(int(job.Jobid)), data)
				Logger.Println(filerr)
			}
			delerr := storedriver.DeleteKeys(redisc, strconv.Itoa(int(job.Jobid)))
			if delerr != nil {
				Logger.Println(delerr)
				if storedriver.CheckActive(redisc) {
					RedisPool.Release(redisc)
				} else {
					redisc.Close()
				}
			} else {
				RedisPool.Release(redisc)
			}

		} else {
			switch job.Jobstatus {
			case 0:
				delete(data, "starttime")
				delete(data, "endtime")
				delete(data, "nodescount")
				delete(data, "cpuscount")
				delete(data, "memory")
				delete(data, "gpus")
				delete(data, "nodes")
				delete(data, "tres_text")

			case 1:
				delete(data, "endtime")
			case 2:
				delete(data, "endtime")
			}
			redisc, er := RedisPool.Acquire()
			if er == nil {
				redierr := storedriver.SetData(redisc, strconv.Itoa(int(job.Jobid)), data)
				if redierr != nil {
					Logger.Println(redierr)
				}
				if storedriver.CheckActive(redisc) {
					RedisPool.Release(redisc)
				} else {
					redisc.Close()
				}

			}

		}

	}

}
func (job *Job) Print() {
	fmt.Printf("jobid: %d,jobname:%s,queue:%s,qtime:%d,starttime:%d,endtime:%d,sub:%s,status:%d,nodes:%d,nodes:%s,cpus:%d,gpus:%d, tres:%s,mem:%d,err:%s,out:%s,command:%s\n",
		job.Jobid, job.Jobname, job.Queue, job.Qtime, job.Starttime, job.Endtime, job.Username, job.Jobstatus, job.Jobnodes, job.Nodes, job.Jobcpus, job.Jobgpus, job.Jobtres, job.Jobmemory, job.Joberrfile, job.Joboutfile, job.Jobcommand)
}

var config *storedriver.MonitorConfig
var configerror error
var rediserror error
var mysqlerror error
var badgerror error
var path *string
var RedisPool *storedriver.Pool
var MysqlPool *gorose.Engin
var Table string
var FilesHandle *storedriver.Badger
var LogPath string = "/var/log/montor.log"
var Logger *log.Logger

func SetCpu(con *storedriver.MonitorConfig) {
	if con.CpuNum > runtime.NumCPU() {
		con.CpuNum = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(con.CpuNum)
}

func GetRedisPool(con *storedriver.MonitorConfig) (*storedriver.Pool, error) {
	redisconfig := &con.Db.Rediscon
	return storedriver.NewPool(redisconfig, 8)

}
func GetMysqlPool(con *storedriver.MonitorConfig) (*gorose.Engin, error) {
	mysqlconfig := &con.Db.Mysqlcon
	return storedriver.Open(mysqlconfig)

}
func GetFilesHandle(con *storedriver.MonitorConfig) (*storedriver.Badger, error) {
	dadgerconfig := &storedriver.Badger{
		ValueDir: con.Db.LocatDir,
	}
	err := dadgerconfig.Open()
	return dadgerconfig, err

}
func SyncBadgerJob() {
	datas := FilesHandle.GetAll()
	for data := range datas {
		item := datas[data]
		err := storedriver.InsertData(MysqlPool, Table, item)
		if err == nil {
			jobid := item["jobid"]
			jobidstr := strconv.Itoa(jobid.(int))
			FilesHandle.Delete(jobidstr)
		}
	}
}
func DelAllFromRedis() {
	redisc, err := RedisPool.Acquire()
	if err == nil {
		storedriver.DeleteKeys(redisc, "*")
		RedisPool.Release(redisc)
	}

}
func init() {
	path = flag.String("c", "/etc/slurm/slurm.yaml", "config path")
	flag.Parse()
	config, configerror = storedriver.ParseConfig(*path)
	if configerror != nil {
		panic(configerror)
	}
	SetCpu(config)
	RedisPool, rediserror = GetRedisPool(config)
	if rediserror != nil {
		panic(rediserror)
	}
	MysqlPool, mysqlerror = GetMysqlPool(config)
	if mysqlerror != nil {
		panic(mysqlerror)
	}
	FilesHandle, badgerror = GetFilesHandle(config)
	if badgerror != nil {
		panic(badgerror)
	}
	Table = config.Db.Mysqlcon.Table
	SyncBadgerJob()
	DelAllFromRedis()
	file, _ := os.OpenFile(LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	Logger = log.New(file, "Trace:", log.LstdFlags|log.Llongfile)

}

type Controller struct {
	Second    int
	Number    int
	Completed chan error
	MaxNumber int
	Mutex     sync.Mutex
	Direct    bool
	FlagCom   bool
}
type Partition struct {
	Start int
	End   int
}

func GetPartition(size, num int) ([]Partition, error) {
	var partitions []Partition
	parmError := errors.New("inpute size and num error")
	if num == 1 {
		if size > 0 {
			p := Partition{Start: 0, End: size}
			partitions = append(partitions, p)
		}
		return partitions, nil
	}
	if size > 0 && num > 0 {
		res := size / num
		i := 1
		for i = 1; i <= num; i++ {
			p := Partition{
				Start: (i - 1) * res,
				End:   i * res,
			}
			partitions = append(partitions, p)

		}
		if num*res < size {
			partitions[num-1].End = size - 1
		}
		return partitions, nil
	}
	return partitions, parmError

}
func HandleOneRecye(con *Controller, direct bool) {
	log.Println("HandleOneRecye")
	con.Mutex.Lock()
	num := con.Number
	con.FlagCom = false
	con.Mutex.Unlock()
	var wg sync.WaitGroup
	jobs := C.get_jobs()
	size := int(jobs.record_count)
	log.Println("size", size)
	if size > 0 {
		if num > size {
			num = size
		}
		wg.Add(num)
		partitions, partitionerror := GetPartition(size, num)

		if partitionerror == nil {
			for item := range partitions {
				partition := partitions[item]
				go HandlePartion(&wg, &partition, jobs, Table)
			}
			wg.Wait()

		}
	}
	con.Mutex.Lock()
	con.FlagCom = true
	con.Mutex.Unlock()
	if !direct {
		con.Completed <- nil
	}
	log.Println("End HandleOneRecye")
}

func HandlePartion(wg *sync.WaitGroup, p *Partition, jobs *C.struct_job_info_msg, table string) {
	log.Println("HandlePartion")
	defer wg.Done()
	if p.End == 0 {
		job := C.get_job(jobs, C.int(0))
		onejob := &Job{}
		onejob.Set(job)
		onejob.Print()
		onejob.Write(table)
	} else {
		log.Println("start:", p.Start)
		log.Println("end:", p.End)
		for item := p.Start; item < p.End; item++ {
			log.Println("index:", item)
			job := C.get_job(jobs, C.int(item))
			onejob := &Job{}
			onejob.Set(job)
			onejob.Print()
			onejob.Write(table)
		}
	}
	log.Println("End HandlePartion")
}

func main() {
	con := &Controller{
		Second:    config.Recyle,
		MaxNumber: config.CpuNum,
		Number:    1,
		Completed: make(chan error),
		Direct:    false,
		FlagCom:   false,
	}
	ticker := time.NewTicker(time.Second * 8)
	for {
		if con.Direct {
			HandleOneRecye(con, true)
		} else {

			go func() {
				HandleOneRecye(con, false)
			}()
			select {
			case <-con.Completed:
				ticker.Stop()
			case <-ticker.C:
				for {
					if con.FlagCom {
						break
					}
				}
				con.Mutex.Lock()
				con.Number = con.Number + 1
				if con.Number > con.MaxNumber {
					con.Direct = true
				}
				con.Mutex.Unlock()

			}
		}
		time.Sleep(3 * time.Second)
		log.Println("this number:", con.Number)
	}

}
