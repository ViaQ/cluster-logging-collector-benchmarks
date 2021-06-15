#!/bin/bash

report_format=${1:-default}
namespace=${2:-logstress}

report=default_report
if [ "${report_format}" == "ndjson" ] ; then
  report=ndjson_report
fi

default_report() {
    echo "====> Top information on: $(date)";
    COLUMNS=1000 top -b -n 1 -o +%CPU | sed 1,6d | awk '{print $9"\t|\t"$10"\t|\t"$12}' | column -t | head -n 20;
    echo "=============";
    echo "====> K8S Top information";
    ./kubectl top pod --namespace="logstress" --sort-by="cpu" --use-protocol-buffers | head -n 20;
    ./kubectl top pod --namespace="loki" --sort-by="cpu" --use-protocol-buffers | head -n 20;
    echo "=============";
}

read -r -d '' SCRIPT << EOM
import sys,datetime,json
a=[]
for l in sys.stdin:
  parts = l.split()
  a.append({"container":parts[0],"cpu":parts[1],"mem":parts[2]})
summary = {"time":datetime.datetime.now().strftime("%Y-%m-%dT%H:%M:%S.%f3")}
stats = {"summary":summary,"resources":a}
print(json.dumps(stats))
EOM

ndjson_report() {
    echo "$(./kubectl top pods --namespace="$namespace" --use-protocol-buffers | head -n 20 | grep -v NAME)" | \
    python3 -c "${SCRIPT}"

}

$report
