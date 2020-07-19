#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <slurm/slurm.h>
#include <slurm/slurm_errno.h>

struct job_resources {
        bitstr_t *core_bitmap;
        bitstr_t *core_bitmap_used;
        uint32_t  cpu_array_cnt;
        uint16_t *cpu_array_value;
        uint32_t *cpu_array_reps;
        uint16_t *cpus;
        uint16_t *cpus_used;
        uint16_t *cores_per_socket;
        uint64_t *memory_allocated;
        uint64_t *memory_used;
        uint32_t  nhosts;
        bitstr_t *node_bitmap;
        uint32_t  node_req;
        char     *nodes;
        uint32_t  ncpus;
        uint32_t *sock_core_rep_count;
        uint16_t *sockets_per_node;
        uint16_t *tasks_per_node;
        uint8_t   whole_node;
};

struct privitejob {
    uint32_t jobid;
    char* name;
    char* partition;
    time_t submit_time;
    time_t start_time;
    time_t end_time;
    char* user_name;
    int job_state;
    int num_nodes;
    int num_cpus;
    char* tres_per_job;
    uint64_t memory_used;
    char* std_err;
    char* outfile;
    char* command;
}

void printall()
{
        int i, j, slurm_err;
        uint64_t mem_alloc, mem_used;
        job_info_msg_t *jobs;

        /* Load job info from Slurm */
        slurm_err = slurm_load_jobs((time_t) NULL, &jobs, SHOW_DETAIL);
        printf("job_id,cluster,partition,user_id,name,job_state,mem_allocated,mem_used\n");
        /* Print jobs info to the file in CSV format */
        for (i = 0; i < jobs->record_count; i++)
        {
                mem_alloc = 0;
                mem_used = 0;
                for (j = 0; j < jobs->job_array[i].job_resrcs->nhosts; j++)
                {
                        mem_alloc += jobs->job_array[i].job_resrcs->memory_allocated[j];
                        mem_used  += jobs->job_array[i].job_resrcs->memory_used[j];
                }
                printf("%d,%s,%s,%d,%s,%d,%d,%d\n",
                        jobs->job_array[i].job_id,
                        jobs->job_array[i].cluster,
                        jobs->job_array[i].partition,
                        jobs->job_array[i].user_id,
                        jobs->job_array[i].name,
                        jobs->job_array[i].job_state,
                        mem_alloc,
                        mem_used
                );
        }
        slurm_free_job_info_msg(jobs);
}

job_info_msg_t *get_jobs(){
   int slurm_err;
   job_info_msg_t *jobs;
   slurm_err = slurm_load_jobs((time_t) NULL, &jobs, SHOW_DETAIL);
   return jobs;   
}

void free_jobs(job_info_msg_t *jobs){
    slurm_free_job_info_msg(jobs); 
}

struct privitejob get_job(job_info_msg_t *jobs, int index){
    slurm_job_info_t slurmjob=NULL;
    struct privitejob job;
    slurmjob =jobs.job_array[index];
    job.jobid = slurmjob.job_id;
    job.name = slurmjob.jobname;
    job.partition = slurmjob.partition;
    job.submit_time = slurmjob.submit_time;
    job.start_time = slurmjob.start_time;
    job.end_time = slurmjob.end_time;
    job.user_name = slurmjob.user_name;
    job.job_state = slurmjob.job_state;
    job.num_nodes = slurmjob.num_nodes;
    job.num_cpus = slurmjob.num_cpus;
    job.tres_per_job = slurmjob.tres_per_job;
    job.std_err = slurmjob.std_err;
    job.std_out = slurmjob.std_out;
    job.command = slurmjob.command;
    if slurmjob.job_resrcs != NULL{
                job.memory_used = 0;
                int j;
                for (j = 0; j < slurmjob.job_resrcs->nhosts; j++)
                {
                         job.memory_used  += slurmjob.job_resrcs->memory_used[j];
                }
        
    }
    return job;
       
}
