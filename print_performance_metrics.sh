# This script collects metrics like log-per-sec, cpu-percentage, cpu-cores and memory required per iterations in running reg-ex rules
# in discovering log-levels.
# collector values could be fluentd, fluentbit
export collector=$1
export capture_pod=`oc get pods | grep capture | cut -d" " -f1`
export  collector_pod=`oc get pods | grep $collector | cut -d" " -f1`
export iterations=`oc logs $capture_pod | grep  "Top information on:" | wc -l`

echo "$collector $collector_pod" 
# if [ "$iterations" -lt 10 ]; then echo "Total Iterations till now: $iterations Results will be printed after 10 iterations"; exit 1; fi
echo "Total Iterations till now: $iterations"

 while : ; do
    iterations=`oc logs $capture_pod | grep  "Top information on:" | wc -l`
		export total_time=`oc logs $capture_pod | grep "Time from start monitoring (in secs)" | cut -d ":" -f2 | tr -d ' ' | tail -1`
    echo "Total Iterations till now: $iterations"
	  echo "Total time till now: $total_time"
		export current_LPS=`oc logs $capture_pod | grep  -i "Total collected logs per sec:" | cut -d ":" -f2 | tr -d ' ' | tail -1`
		echo "Current LPS=$current_LPS"
    export LPS=`oc logs $capture_pod | grep  -i "Total collected logs per sec:" | cut -d ":" -f2 | awk '{ SUM += $1} END { print SUM/NR }'`
    echo "Avg logs per sec/iter: $LPS"

		export current_cpu_core=`oc logs $capture_pod   | grep $collector_pod  |  awk '{print $2}' | cut -d 'm' -f1 | tail -1`
		echo "Current CPU core=$current_cpu_core"
    export Cpu_Core=`oc logs $capture_pod   | grep $collector_pod  |  awk '{print $2}' | cut -d 'm' -f1  | awk '{ SUM+= $1} END { print SUM/NR }'`
    echo "Avg cpu core/iter: $Cpu_Core"
		export current_memory=`oc logs $capture_pod  | grep $collector_pod  |  awk '{print $3}' | cut -d 'M' -f1 | tail -1`
		echo "Current Memory=$current_memory"
    export Memory=`oc logs $capture_pod  | grep $collector_pod  |  awk '{print $3}' | cut -d 'M' -f1  | awk '{ SUM+= $1} END { print SUM/NR }'`
    echo "Avg memory/iter: $Memory"
    export end_time=$(date +%s%N | cut -b1-13)
    echo "End time: $end_time"
    if [ "$total_time" -ge 1800 ]; then break; fi
    sleep 10
  done

export end_time=$(date +%s%N | cut -b1-13)
echo "Start time: $start_time"
echo "End time: $end_time"
echo "Exiting after $iterations"
