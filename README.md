#  Cluster Logging Collectors - benchmarks

This repository holds code and scripts to benchmark
fluentd, fluent-bit and vector log collectors as they are deployed 
on top of  environment.

The content of this repository includes the following components:

1. Log load application deployed inside containers that create log stress on OpenShift. (Please refer to [cluster-logging-load-client](https://github.com/ViaQ/cluster-logging-load-client) for more information) 
1. Collectors configuration and deployment files. This includes fluentd, fluent bit and vector configurations   
1. Benchmark monitoring and statistics component (simple golang app `check-logs-sequence.go`)   
1. Deployment and benchmark management scripts  

> Note: The benchmark intentionally uses single OpenShift worker node. The benchmark script
will choose one of the available nodes and deploy all benchmark components
onto that node

> Note: to maximize benchmark accuracy, it is highly recommended that 
the used cluster is not deployed with any additional 
workloads/containers other than the ones deployed by this benchmark. 
If required, it is possible to use  **evacuation** configuration
command-line parameter to evacuate all none-related pods from the used node, but this is less recommended 

## Prerequisites

- Deployed OCP cluster
- Golang version 1.5.2 or higher

## Installation

1. Login to your  cluster (using `oc login` command)
1. Clone this repository
1. Execute  
`./deploy_to_openshift.sh`

## Configuration

For complete list of configuration 
options execute: `./deploy_to_openshift.sh --help`

For example to benchmark using **fluentbit** as the log collector
and use **heavy** configuration profile [(detailed profile parameters)](https://github.com/ViaQ/cluster-logging-collector-benchmarks/pull/1/files#diff-44133797f573b7ceda048bb2dc56353ef30a40de72ffdfb7afc6cd5754d77339R84)
execute   
`./deploy_to_openshift.sh -c=fluentbit -p=heavy`

Another example using fluentd with specific configuration (partial CLO configuration)  
`./deploy_to_openshift.sh -p=heavy -c=fluentd -fc=conf/collector/fluentd/partial/CLO_no_measure.conf`

> Note: make sure to use the sign `=` between each command-line key and value

### Cluster Logging Load Client

This [project](https://github.com/ViaQ/cluster-logging-load-client) is a golang application to generate logs and send them to various output destinations in various formats. The app runs as a single executable and based on configuration it can spawn multiple threads. User can scale the app horizontally for heavy workload. 

To use a different load client, you can specify it in the configuration  
`./deploy_to_openshift.sh --stressorimage=<LINK TO YOUR CONTAINER IMAGE>`


### Vector Configuration
Note any vector configuration that utilizes a "kubernetes_log" source must additionally have a transform to modify "file" to "path" as
in the default configuration.  Capturing statistics will otherwise not measure log information.

## Typical deployment

Typical deployment of the benchmark components on OpenShift cluster looks like this: 

```
$ oc get pods
NAME                                 READY   STATUS    RESTARTS   AGE
capturestatistics-86cbb9d84d-jgh2d   1/1     Running   0          36s
fluentd-597f957d6b-gmxlb             1/1     Running   0          48s
heavy-log-stress-66dd57fb95-ghjv2    1/1     Running   0          56s
heavy-log-stress-66dd57fb95-kz95g    1/1     Running   0          56s
low-log-stress-6db87fbcbc-j5l8s      1/1     Running   0          56s
low-log-stress-6db87fbcbc-l57sb      1/1     Running   0          56s
low-log-stress-6db87fbcbc-lh5s2      1/1     Running   0          56s
low-log-stress-6db87fbcbc-llj84      1/1     Running   0          55s
low-log-stress-6db87fbcbc-nws4l      1/1     Running   0          56s
low-log-stress-6db87fbcbc-pljv4      1/1     Running   0          56s
low-log-stress-6db87fbcbc-r79vv      1/1     Running   0          56s
low-log-stress-6db87fbcbc-x4w8p      1/1     Running   0          55s
```

## Benchmark results

Benchmark results are logged periodically onto **capturestatistics** pod and include: both 
1. `top` information every 120 seconds (e.g. cpu, memory )
1. Statistical information on logs created by **stress** containers 
   and log captured by the collector. The information is available periodically and as total from
   beginning of benchmark. In addition, log loss is calculated

```
====> Top information on: Mon Mar 15 13:53:14 UTC 2021
top - 13:53:14 up  8:25,  0 users,  load average: 0.51, 0.62, 2.22
Tasks: 438 total,   1 running, 437 sleeping,   0 stopped,   0 zombie
%Cpu(s):  1.1 us,  1.1 sy,  0.0 ni, 97.8 id,  0.0 wa,  0.0 hi,  0.0 si,  0.0 st
MiB Mem :  63605.7 total,  41370.7 free,   5408.4 used,  16826.6 buff/cache
MiB Swap:      0.0 total,      0.0 free,      0.0 used.  58234.2 avail Mem

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
   1876 root      20   0 3415160 192296  66372 S  12.5   0.3  61:22.61 kubelet
   8403 nobody    20   0  739148  36756  22240 S   6.2   0.1   0:04.47 kube-rbac-proxy
   8459 nobody    20   0 3380228   2.0g 158176 S   6.2   3.3  61:41.15 prometheus
 508900 root      20   0  143732   2704   1976 S   6.2   0.0   0:04.97 conmon
 509792 root      20   0  110996  40860   8996 S   6.2   0.1   3:08.95 fluent-bit
 532842 root      20   0    7696   4084   3284 R   6.2   0.0   0:00.01 top
      1 root      20   0  249324  17652   9036 S   0.0   0.0   3:23.92 systemd
      2 root      20   0       0      0      0 S   0.0   0.0   0:00.05 kthreadd
      3 root       0 -20       0      0      0 I   0.0   0.0   0:00.00 rcu_gp
      4 root       0 -20       0      0      0 I   0.0   0.0   0:00.00 rcu_par_gp
      6 root       0 -20       0      0      0 I   0.0   0.0   0:00.00 kworker/0:0H-kblockd
      9 root       0 -20       0      0      0 I   0.0   0.0   0:00.00 mm_percpu_wq
     10 root      20   0       0      0      0 S   0.0   0.0   0:00.81 ksoftirqd/0
     11 root      20   0       0      0      0 I   0.0   0.0   0:12.28 rcu_sched
     12 root      rt   0       0      0      0 S   0.0   0.0   0:00.03 migration/0
     13 root      rt   0       0      0      0 S   0.0   0.0   0:00.00 watchdog/0
     14 root      20   0       0      0      0 S   0.0   0.0   0:00.00 cpuhp/0
     15 root      20   0       0      0      0 S   0.0   0.0   0:00.01 cpuhp/1
     16 root      rt   0       0      0      0 S   0.0   0.0   0:00.02 watchdog/1
     17 root      rt   0       0      0      0 S   0.0   0.0   0:00.03 migration/1
     18 root      20   0       0      0      0 S   0.0   0.0   0:00.48 ksoftirqd/1
     20 root       0 -20       0      0      0 I   0.0   0.0   0:00.00 kworker/1:0H-kblockd
     21 root      20   0       0      0      0 S   0.0   0.0   0:00.01 cpuhp/2
     22 root      rt   0       0      0      0 S   0.0   0.0   0:00.02 watchdog/2
     23 root      rt   0       0      0      0 S   0.0   0.0   0:00.02 migration/2
     24 root      20   0       0      0      0 S   0.0   0.0   0:00.38 ksoftirqd/2
     26 root       0 -20       0      0      0 I   0.0   0.0   0:00.00 kworker/2:0H-kblockd
     27 root      20   0       0      0      0 S   0.0   0.0   0:00.01 cpuhp/3
     28 root      rt   0       0      0      0 S   0.0   0.0   0:00.01 watchdog/3
     29 root      rt   0       0      0      0 S   0.0   0.0   0:00.03 migration/3
     30 root      20   0       0      0      0 S   0.0   0.0   0:01.03 ksoftirqd/3
     32 root       0 -20       0      0      0 I   0.0   0.0   0:00.00 kworker/3:0H-kblockd
     33 root      20   0       0      0      0 S   0.0   0.0   0:00.01 cpuhp/4
2021/03/15 13:53:15 Report at: 2021-03-15 13:53:15.789149368 +0000 UTC m=+1443.880773817
2021/03/15 13:53:15 -==-=-=-=-=
2021/03/15 13:53:15 Total number of collected logs : 21600000
2021/03/15 13:53:15 Logs per sec : 14958
2021/03/15 13:53:15 Time from start monitoring (in secs): 1444
2021/03/15 13:53:15 -==-=-=-=-=
2021/03/15 13:53:15 ----------------------------------------------------------------------------------------------------------------------------------------
2021/03/15 13:53:15 |                                      | Current   | Lines     |           | Total     | Lines     |           |           |           |
2021/03/15 13:53:15 ----------------------------------------------------------------------------------------------------------------------------------------
2021/03/15 13:53:15 | Container name                       | Logged    | Collected | Loss      | Logged    | Lo./Sec   | Collected | Co./Sec   | Loss      |
2021/03/15 13:53:15 ----------------------------------------------------------------------------------------------------------------------------------------
2021/03/15 13:53:15 | low-log-stress-7c7f49566-49l24       | 19950     | 19950     | 0         | 2160000   | 1495      | 2159700   | 1495      | 300       |
2021/03/15 13:53:15 | low-log-stress-7c7f49566-6rsl2       | 19950     | 19950     | 0         | 2160150   | 1495      | 2160000   | 1495      | 150       |
2021/03/15 13:53:15 | low-log-stress-7c7f49566-6w4gk       | 20035     | 20035     | 0         | 2160085   | 1495      | 2159935   | 1495      | 150       |
2021/03/15 13:53:15 | low-log-stress-7c7f49566-8bwsv       | 19997     | 19997     | 0         | 2160000   | 1495      | 2159999   | 1495      | 1         |
2021/03/15 13:53:15 | low-log-stress-7c7f49566-8vl2j       | 19950     | 19950     | 0         | 2160000   | 1495      | 2159850   | 1495      | 150       |
2021/03/15 13:53:15 | low-log-stress-7c7f49566-c2w56       | 20035     | 20035     | 0         | 2160085   | 1495      | 2160084   | 1495      | 1         |
2021/03/15 13:53:15 | low-log-stress-7c7f49566-f5wp9       | 19998     | 19998     | 0         | 2160198   | 1495      | 2160197   | 1495      | 1         |
2021/03/15 13:53:15 | low-log-stress-7c7f49566-r4c6v       | 20185     | 20185     | 0         | 2160385   | 1496      | 2160385   | 1496      | 0         |
2021/03/15 13:53:15 | low-log-stress-7c7f49566-rmz24       | 19950     | 19950     | 0         | 2160000   | 1495      | 2160000   | 1495      | 0         |
2021/03/15 13:53:15 | low-log-stress-7c7f49566-vxfmp       | 19950     | 19950     | 0         | 2160000   | 1495      | 2159850   | 1495      | 150       |
2021/03/15 13:53:15

```

## Automation

Running the benchmark against various scenarios can be achieved using the `auto_execution.sh` script for example::

```bash
./contrib/auto_execution.sh -ff=conf/collector/fluentd/partial/
```

For more details execute with `-h` parameter





