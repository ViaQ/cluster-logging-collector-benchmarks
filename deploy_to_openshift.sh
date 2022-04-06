#!/bin/bash

show_usage() {
  echo "
usage: deploy_to_openshift [options]
  options:
    -h,  --help                Show usage
    -e,  --evacuate=[enum]     Evacuate node  (false, true  default: false)
    -p,  --profile=[enum]      Stress profile (no-stress, very-light, light, medium, heavy, very-heavy, heavy-loss, very-heavy-loss  default: very-light)
    -c,  --collector=[enum]    Logs collector (fluentd, fluentbit, fluentd-loki, fluentbit-loki, dual, vector default: fluentd)
    -fc, --fluentconf=[enum]   fluent conf file (default: conf/collector/fluentd/fluentd-clf.conf)
    -si, --stressorimage=[string] Log Stressor image to use (default: quay.io/openshift-logging/cluster-logging-load-client:latest)
    -fi, --fimage=[string]     Fluentd image to use (default: quay.io/openshift/origin-logging-fluentd:latest)
    -bi, --bimage=[string]     Fluentbit image to use (default: quay.io/openshift/origin-logging-fluentd:latest)
    -bc, --bconf=[string]      Fluentbit config to use (default: conf/collector/fluentbit/fluentbit-clf.conf)
    -vi, --vimage=[string]     Vector image to use (default: quay.io/dciancio/vector:latest)
    -vc, --vectorconf=[string] Vector config to use (default: conf/collector/vector/vector-clf.toml)
    -gi, --gimage=[string]     Gologfilewatcher image to use (default: docker.io/cognetive/go-log-file-watcher-driver-v0)
    -gw, --gowatcher=[enum]    Deploy go watcher  (false, true  default: false)
    -lf, --use_log_samples=[enum]   Use log samples to generate workload logs  (false, true  default: false)
    -fp, --fluentd_pre=[path]  shell script to execute before fluentd starts (default =\"\")
    -of, --output_format=[enum] Output format: (ndjson, default: default)
    -ri, --report_interval     Report Interval (default: 120)
"
  exit 0
}

select_stress_profile() {
  number_heavy_stress_containers=2
  number_low_stress_containers=10
  heavy_containers_msg_per_sec=1000
  low_containers_msg_per_sec=10
  number_of_log_lines_between_reports=10;
  maximum_logfile_size=10485760;

  case $stress_profile in
      "benchmarking")
        number_heavy_stress_containers=0;
        heavy_containers_msg_per_sec=0;
        number_low_stress_containers=200;
        low_containers_msg_per_sec=50;
        number_of_log_lines_between_reports=10;
        maximum_logfile_size=51200000;
        ;;
      "no-stress")
        number_heavy_stress_containers=0;
        heavy_containers_msg_per_sec=0;
        number_low_stress_containers=0;
        low_containers_msg_per_sec=0;
        number_of_log_lines_between_reports=10;
        maximum_logfile_size=10485760;
        ;;
      "very-light")
        number_heavy_stress_containers=0;
        heavy_containers_msg_per_sec=0;
        number_low_stress_containers=1;
        low_containers_msg_per_sec=10;
        number_of_log_lines_between_reports=100;
        maximum_logfile_size=10485760;
        ;;
      "light")
        number_heavy_stress_containers=1;
        heavy_containers_msg_per_sec=100;
        number_low_stress_containers=2;
        low_containers_msg_per_sec=10;
        number_of_log_lines_between_reports=1000;
        maximum_logfile_size=1048576;
        ;;
      "experiment")
        number_heavy_stress_containers=0;
        heavy_containers_msg_per_sec=0;
        number_low_stress_containers=20;
        low_containers_msg_per_sec=20000;
        number_of_log_lines_between_reports=100;
        maximum_logfile_size=10485760;
        ;;
      "medium")
        number_heavy_stress_containers=2;
        heavy_containers_msg_per_sec=1000;
        number_low_stress_containers=10;
        low_containers_msg_per_sec=10;
        number_of_log_lines_between_reports=20000;
        maximum_logfile_size=1048576;
        ;;
      "heavy")
        number_heavy_stress_containers=0;
        heavy_containers_msg_per_sec=0;
        number_low_stress_containers=10;
        low_containers_msg_per_sec=1500;
        number_of_log_lines_between_reports=200000;
        maximum_logfile_size=1048576;
        ;;
      "very-heavy")
        number_heavy_stress_containers=0;
        heavy_containers_msg_per_sec=0;
        number_low_stress_containers=10;
        low_containers_msg_per_sec=3000;
        number_of_log_lines_between_reports=300000;
        maximum_logfile_size=1048576;
        ;;
      "heavy-loss")
        number_heavy_stress_containers=2;
        heavy_containers_msg_per_sec=20000;
        number_low_stress_containers=8;
        low_containers_msg_per_sec=1500;
        number_of_log_lines_between_reports=200000;
        maximum_logfile_size=1048576;
        ;;
      "very-heavy-loss")
        number_heavy_stress_containers=10;
        heavy_containers_msg_per_sec=20000;
        number_low_stress_containers=10;
        low_containers_msg_per_sec=1500;
        number_of_log_lines_between_reports=1000000;
        maximum_logfile_size=1048576;
        ;;
      *) show_usage
        ;;
  esac
}

show_configuration() {

echo "
Note: get more deployment options with -h

Configuration:
-=-=-=-=-=-=-
Evacuate node --> $evacuate_node
Stress profile --> $stress_profile
Logs collector --> $collector

Log stressor image --> $stressor_image
number of heavy stress containers --> $number_heavy_stress_containers
Heavy stress containers msg per second --> $heavy_containers_msg_per_sec
number of low stress containers --> $number_low_stress_containers
Low stress containers msg per second --> $low_containers_msg_per_sec

Number of log lines between reports --> $number_of_log_lines_between_reports
Maximum size of log file --> $maximum_logfile_size

Deploy gowatcher --> $gowatcher
Gologfilewatcher image used --> $gologfilewatcher_image
Use logsample file --> $use_log_samples
Output Format --> $output_format

Fluentd image used --> $fluentd_image
Fluentd conf used --> $fluentd_conf_file
Fluentd pre script --> $fluentd_pre

Fluentbit image used --> $fluentbit_image
Fluentbit conf used --> $fluentbit_conf

Vector image used --> $vector_image
Vector conf used --> $vector_conf

"
}

deploy() {
  select_stress_profile
  show_configuration
  select_node_to_use
  configure_workers_log_rotation $maximum_logfile_size
  return_node_to_normal
  delete_logstress_project_if_exists
  create_logstress_project
  set_credentials
  deploy_logstress $number_heavy_stress_containers $heavy_containers_msg_per_sec $number_low_stress_containers $low_containers_msg_per_sec $use_log_samples $stressor_image
  if $gowatcher ; then deploy_gologfilewatcher "$gologfilewatcher_image"; fi
  deploy_capture_statistics $number_of_log_lines_between_reports "$output_format" "$report_interval"
	sleep 10
  case "$collector" in
    'vector') deploy_log_collector_vector "$vector_image" "$vector_conf";;
    'fluentd') deploy_log_collector_fluentd "$fluentd_image" "$fluentd_conf_file" "$fluentd_pre";;
    'fluentbit') deploy_log_collector_fluentbit "$fluentbit_image" "$fluentbit_conf";;
    'fluentd-loki') deploy_log_collector_fluentd "$fluentd_image" conf/collector/fluentd/loki/fluentd.conf "$fluentd_pre";;
    'fluentbit-loki') deploy_log_collector_fluentbit "$fluentbit_image" conf/collector/fluentbit/loki/fluentbit.conf;;
    'dual') deploy_log_collector_fluentd "$fluentd_image" conf/collector/fluentd/dual/fluentd.conf "$fluentd_pre";
            deploy_log_collector_fluentbit "$fluentbit_image" conf/collector/fluentbit/dual/fluentbit.conf;;
    *) show_usage ;;
  esac
  if $gowatcher ; then expose_metrics_to_prometheus; fi
  if $evacuate_node ; then evacuate_node_for_performance_tests; fi
}


RUNNING="$(basename $(echo "$0" | sed 's/-//g'))"
if [[ "$RUNNING" == "deploy_to_openshift.sh" ]]
then

  source ./contrib/deploy_functions.sh
  source ./contrib/expose_metrics.sh

  #default parameters
  stress_profile="very-light"
  evacuate_node="false"
  stressor_image="quay.io/openshift-logging/cluster-logging-load-client:latest"
  fluentd_image="docker.io/cognetive/origin-logging-fluentd:0.1"
  fluentd_conf_file="conf/collector/fluentd/fluentd-clf.conf"
  gologfilewatcher_image="docker.io/cognetive/go-log-file-watcher-with-symlink-support-v0"
  fluentbit_image="fluent/fluent-bit:1.7-debug"
  fluentbit_conf="conf/collector/fluentbit/fluentbit-clf.conf"
  vector_image="timberio/vector:latest-alpine"
  #vector_image=quay.io/openshift-logging/vector:0.14.1
  vector_conf="conf/collector/vector/vector-clf.toml"
  collector="fluentd"
  fluentd_pre=""
  gowatcher="false"
  use_log_samples="false"
  output_format="default"
  report_interval="120"

  for i in "$@"
  do
  case $i in
      -e=*|--evacuate_node=*) evacuate_node="${i#*=}"; shift ;;
      -p=*|--profile=*) stress_profile="${i#*=}"; shift ;;
      -c=*|--collector=*) collector="${i#*=}"; shift ;;
      -fc=*|--fluentconf=*) fluent_conf_file="${i#*=}"; shift ;;
      -si=*|--stressorimage=*) stressor_image="${i#*=}"; shift ;;
      -fi=*|--fimage=*) fluentd_image="${i#*=}"; shift ;;
      -bi=*|--bimage=*) fluentbit_image="${i#*=}"; shift ;;
      -bc=*|--bconf=*) fluentbit_conf="${i#*=}"; shift ;;
      -vi=*|--vimage=*) vector_image="${i#*=}"; shift ;;
      -vc=*|--vectorconf=*) vector_conf="${i#*=}"; shift ;;
      -gw=*|--gowatcher=*) gowatcher="${i#*=}"; shift ;;
      -lf=*|--use_log_samples=*) use_log_samples="${i#*=}"; shift ;;
      -gi=*|--gimage=*) gologfilewatcher_image="${i#*=}"; shift ;;
      -fp=*|--fluentd_pre=*) fluentd_pre="${i#*=}"; shift ;;
      -of=*|--output_format=*) output_format="${i#*=}"; shift ;;
      -ri=*|--report_interval=*) report_interval="${i#*=}"; shift ;;
      --nothing) nothing=true; shift ;;
      -h|--help|*) show_usage ;;
  esac
  done

  if [[ "$use_log_samples" == "true" ]]
  then
    use_log_samples="application"
  else 
    use_log_samples="synthetic"
  fi

  deploy "$@"
  print_pods_status
  print_usage_instructions
fi



