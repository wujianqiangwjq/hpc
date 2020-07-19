package main

/*
#cgo LDFLAGS: -lslurm
#include "slurmclient.h"
*/
import "C"
import (
	"fmt"
)

type Job struct {
	Jobid      uint32
	Jobname    string
	Queue      string
	Qtime      int64
	Starttime  int64
	Endtime    int64
	Submiter   string
	Jobstatus  int
	Jobnodes   int
	Jobcpus    int
	Jobtres    string
	Jobmemory  uint64
	Joberrfile string
	Joboutfile string
	Jobcommand string
}

func (job *Job) Set(pjob C.struct_privitejob) {
	job.Jobid = pjob.jobid
	job.Jobname = C.GoString(pjob.name)
	job.Queue = C.GoString(pjob.partition)
	job.Qtime = int64(pjob.submit_time)
	job.Starttime = int64(pjob.start_time)
	job.Endtime = int64(pjob.end_time)
	job.Submiter = C.GoString(pjob.user_name)
	job.Jobstatus = int(pjob.job_state)
	job.Jobnodes = int(pjob.num_nodes)
	job.Jobcpus = int(pjob.num_cpus)
	job.Jobtres = C.GoString(pjob.tres_per_job)
	job.Jobmemory = uint64(pjob.memory_used)
	job.Joberrfile = C.GoString(pjob.std_err)
	job.Joboutfile = C.GoString(pjob.outfile)
	job.Jobcommand = C.GoString(pjob.command)

}
func (job *Job) Print() {
	fmt.Printf("jobid: %d,jobname:%s,queue:%s,qtime:%d,starttime:%d,endtime:%d,sub:%s,status:%d,nodes:%d,cpus:%d,tres:%s,mem:%d,command:%s,err:%s,out:%s\n",
		job.Jobid, job.Jobname, job.Queue, job.Qtime, job.Starttime, job.Endtime, job.Submiter, job.Jobstatus, job.Jobnodes, job.Jobcpus, job.Jobtres, job.Jobmemory, job.Joberrfile, job.Joboutfile, job.Jobcommand)
}

func main() {
	jobs := C.get_jobs()
	size := uint32(jobs.record_count)
	for index := 0; index < size; index++ {
		job := C.get_job(jobs, C.int(index))
		onejob := &Job{}
		onejob.Set(job)
		onejob.Print()

	}

}
