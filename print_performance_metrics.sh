# This script collects metrics like log-per-sec, cpu-percentage, cpu-cores and memory required per iterations in running reg-ex rules
# in discovering log-levels.

export capture_pod=`oc get pods | grep capture | cut -d" " -f1`
export collector_pod=`oc get pods | grep fluent | cut -d" " -f1`
export iterations=`oc logs $capture_pod | grep  " |  ruby" | wc -l`

if [ "$iterations" -lt 10 ]; then echo "Total Iterations till now: $iterations Results will be printed after 10 iterations"; exit 1; fi

echo "Total No of iterations:  $iterations"

export LPS=`oc logs $capture_pod | grep  -i "Total collected logs per sec:" | cut -d ":" -f2 | awk '{ SUM += $1} END { print SUM/NR }'`
echo "Avg logs per sec/iter: $LPS"

export Cpu_Percentage=`oc logs $capture_pod  | grep -i "|  ruby" | cut -d "|" -f1 | awk '{ SUM += $1} END { print SUM/NR }'`
echo "Avg cpu percentage/iter: $Cpu_Percentage"

export Cpu_Core=`oc logs $capture_pod   | grep $fluentd_pod  |  awk '{print $2}' | cut -d 'm' -f1  | awk '{ SUM+= $1} END { print SUM/NR }'`
echo "Avg cpu core/iter: $Cpu_Core"

export Memory=`oc logs $capture_pod  | grep $fluentd_pod  |  awk '{print $3}' | cut -d 'M' -f1  | awk '{ SUM+= $1} END { print SUM/NR }'`
echo "Avg memory/iter: $Memory"
