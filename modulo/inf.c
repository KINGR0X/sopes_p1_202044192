#include <linux/module.h>
#include <linux/kernel.h>
#include <linux/string.h> 
#include <linux/init.h>
#include <linux/proc_fs.h> 
#include <linux/seq_file.h> 
#include <linux/mm.h> 
#include <linux/sched.h> 
#include <linux/timer.h> 
#include <linux/jiffies.h> 
#include <linux/uaccess.h>
#include <linux/tty.h>
#include <linux/sched/signal.h>
#include <linux/fs.h>        
#include <linux/slab.h>      
#include <linux/sched/mm.h>
#include <linux/binfmts.h>
#include <linux/timekeeping.h>
#include <linux/blkdev.h>
#include <linux/fs_struct.h>
#include <linux/nsproxy.h>
#include <linux/ns_common.h>
#include <linux/pid_namespace.h>

MODULE_LICENSE("GPL");
MODULE_AUTHOR("Elian");
MODULE_DESCRIPTION("Modulo para leer informacion de RAM y CPU, e información de contenedores");
MODULE_VERSION("0.1");

#define PROC_NAME "sysinfo_202044192" 
#define MAX_CMDLINE_LENGTH 256

static char *get_process_cmdline(struct task_struct *task) {
    struct mm_struct *mm;
    char *cmdline, *p;
    unsigned long arg_start, arg_end, env_start;
    int i, len;

    cmdline = kmalloc(MAX_CMDLINE_LENGTH, GFP_KERNEL);
    if (!cmdline) return NULL;

    mm = get_task_mm(task);
    if (!mm) {
        kfree(cmdline);
        return NULL;
    }

    down_read(&mm->mmap_lock);
    arg_start = mm->arg_start;
    arg_end = mm->arg_end;
    env_start = mm->env_start;
    up_read(&mm->mmap_lock);

    len = arg_end - arg_start;
    if (len > MAX_CMDLINE_LENGTH - 1) len = MAX_CMDLINE_LENGTH - 1;

    if (access_process_vm(task, arg_start, cmdline, len, 0) != len) {
        mmput(mm);
        kfree(cmdline);
        return NULL;
    }

    cmdline[len] = '\0';
    p = cmdline;

    for (i = 0; i < len; i++) {
        if (p[i] == '\0') p[i] = ' ';
    }

    mmput(mm);
    return cmdline;
}

static int sysinfo_show(struct seq_file *m, void *v) {
    struct sysinfo si;
    struct task_struct *task;
    unsigned long total_jiffies = jiffies;
    unsigned long total_cpu_time = 0;
    unsigned long total_cpu_usage = 0;
    int first_process = 1;

    si_meminfo(&si);

    // Cálculos para memoria en KB y MB
    unsigned long totalram_kb = si.totalram * (PAGE_SIZE / 1024);
    unsigned long freeram_kb = si.freeram * (PAGE_SIZE / 1024);
    unsigned long ram_usage_kb = totalram_kb - freeram_kb;

    unsigned long totalram_mb = totalram_kb / 1024;
    unsigned long freeram_mb = freeram_kb / 1024;
    unsigned long ram_usage_mb = ram_usage_kb / 1024;

    seq_printf(m, "{\n");
    
    // Información de la memoria en MB
    seq_printf(m, "\"ram_total_mb\": %lu, \n", totalram_mb);
    seq_printf(m, "\"ram_free_mb\": %lu, \n", freeram_mb);
    seq_printf(m, "\"ram_usage_mb\": %lu, \n", ram_usage_mb);

    // Calcular el uso total del CPU
    for_each_process(task) {
        total_cpu_time += task->utime + task->stime;
    }

    total_cpu_usage = (total_cpu_time * 10000) / total_jiffies;

    seq_printf(m, "\"cpu_usage\": %lu.%02lu, \n", total_cpu_usage / 100, total_cpu_usage % 100);

    seq_printf(m, "\"processes\": [\n");

    // Información de los procesos relacionados con contenedores
    for_each_process(task) {
        if (strcmp(task->comm, "containerd-shim") == 0) {
            unsigned long rss = 0;
            unsigned long mem_usage = 0;
            unsigned long cpu_usage = 0;
            char *cmdline = NULL;
            unsigned long long read_bytes = 0, write_bytes = 0;
            struct task_struct *child_task;
            struct pid_namespace *pid_ns = task_active_pid_ns(task);

            if (task->mm) {
                rss = get_mm_rss(task->mm) << (PAGE_SHIFT - 10); // RSS en KB
                mem_usage = (rss * 10000) / totalram_kb; // Uso de memoria en porcentaje
            }

            unsigned long total_time = task->utime + task->stime;
            cpu_usage = (total_time * 10000) / total_jiffies;
            cmdline = get_process_cmdline(task);

            // Obtener información de I/O
            for_each_thread(task, child_task) {
                read_bytes += child_task->ioac.read_bytes;
                write_bytes += child_task->ioac.write_bytes;
            }

            if (!first_process) {
                seq_printf(m, ",\n");
            } else {
                first_process = 0;
            }

            seq_printf(m, "  {\n");
            seq_printf(m, "    \"PID\": %d,\n", task->pid);
            seq_printf(m, "    \"Name\": \"%s\",\n", task->comm);
            seq_printf(m, "    \"ContainerID\": \"%d\",\n", pid_ns->ns.inum);
            seq_printf(m, "    \"MemoryUsage\": %lu.%02lu,\n", mem_usage / 100, mem_usage % 100);
            seq_printf(m, "    \"CPUUsage\": %lu.%02lu,\n", cpu_usage / 100, cpu_usage % 100);
            seq_printf(m, "    \"DiskUsageRead_mb\": %llu,\n", read_bytes / (1024 * 1024));
            seq_printf(m, "    \"DiskUsageWrite_mb\": %llu,\n", write_bytes / (1024 * 1024));
            seq_printf(m, "    \"IORead_mb\": %llu,\n", read_bytes / (1024 * 1024));
            seq_printf(m, "    \"IOWrite_mb\": %llu\n", write_bytes / (1024 * 1024));
            seq_printf(m, "  }");

            if (cmdline) {
                kfree(cmdline);
            }
        }
    }

    seq_printf(m, "\n]\n}\n");
    return 0;
}

static int sysinfo_open(struct inode *inode, struct file *file) {
    return single_open(file, sysinfo_show, NULL);
}

static const struct proc_ops sysinfo_ops = {
    .proc_open = sysinfo_open,
    .proc_read = seq_read,
};

static int __init sysinfo_init(void) {
    proc_create(PROC_NAME, 0, NULL, &sysinfo_ops);
    printk(KERN_INFO "Modulo sysinfo cargado \n");
    return 0;
}

static void __exit sysinfo_exit(void) {
    remove_proc_entry(PROC_NAME, NULL);
    printk(KERN_INFO "Modulo sysinfo desinstalado\n");
}

module_init(sysinfo_init);
module_exit(sysinfo_exit);