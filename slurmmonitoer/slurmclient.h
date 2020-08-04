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
    uint32_t  user_id;
    int job_state;
    int num_nodes;
    int num_cpus;
    char *nodes;
    char* tres_per_node;
    uint64_t memory_used;
    char* std_err;
    char* std_out;
    char* command;
    uint32_t gpus;
    char* work_dir;
};

struct privitenode {
    uint16_t cpus;
    uint32_t cpu_load;
    uint64_t free_mem;
    char* gres;
    char* gres_used;
    char* name;
    uint64_t real_memory;
    uint16_t alloc_cpus;
    uint64_t alloc_memory;
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
   if (slurm_err == -1){
    return NULL;
   }
   return jobs;   
}

void free_jobs(job_info_msg_t *jobs){
    slurm_free_job_info_msg(jobs); 
}

struct privitejob get_job(job_info_msg_t *jobs, int index){
    slurm_job_info_t slurmjob;
    struct privitejob job;
    slurmjob =jobs->job_array[index];
    job.jobid = slurmjob.job_id;
    job.name = slurmjob.name;
    job.partition = slurmjob.partition;
    job.submit_time = slurmjob.submit_time;
    job.start_time = slurmjob.start_time;
    job.end_time = slurmjob.end_time;
    job.user_id = slurmjob.user_id;
    job.job_state = slurmjob.job_state;
    job.num_nodes = slurmjob.num_nodes;
    job.num_cpus = slurmjob.num_cpus;
    job.tres_per_node = slurmjob.tres_per_node;
    job.std_err = slurmjob.std_err;
     job.nodes = slurmjob.nodes;
    job.std_out = slurmjob.std_out;
    job.command = slurmjob.command;
    job.gpus = slurmjob.gres_detail_cnt;
    job.work_dir = slurmjob.work_dir;
    if( slurmjob.job_resrcs != NULL){
                job.memory_used = 0;
                int j;
                for (j = 0; j < slurmjob.job_resrcs->nhosts; j++)
                {
                         job.memory_used  += slurmjob.job_resrcs->memory_used[j];
                }
        
    }
    return job;
       
}


node_info_msg_t *get_nodes(){
    int slurm_err;
    node_info_msg_t *nodes;
    slurm_err = slurm_load_node((time_t) NULL, &nodes, SHOW_DETAIL);
    if (slurm_err == -1){
        return NULL;
    }else{
        return nodes;
    }  
}

struc privitenode get_node(node_info_msg_t *nodes, int index){
    struct privitenode node;
    node_info_t * node_info=NULL;
    node_info = nodes->node_array[index];
    node.name = node_info.name;
    node.cpus = node_info.cpus;
    node.cpu_load = node_info.cpu_load;
    node.free_mem =node_info.free_mem;
    node.gres = node_info.gres;
    node.gres_used = node_info.gres_used;
    node.real_memory = node_info.real_memory;
    slurm_get_select_nodeinfo(node_info.select_nodeinfo, SELECT_NODEDATA_SUBCNT,NODE_STATE_ALLOCATED,&node.alloc_cpus);
    slurm_get_select_nodeinfo(node_info.select_nodeinfo, SELECT_NODEDATA_MEM_ALLOC,NODE_STATE_ALLOCATED,&node.alloc_memory);
    return node;
}

void free_nodes(node_info_msg_t * nodes){
    slurm_free_node_info_msg(nodes);
}
