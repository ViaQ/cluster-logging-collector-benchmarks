#!/bin/bash

auto_show_usage() {
  echo "
usage: auto_execution [options]
  options:

    -h,  --help                           Show usage
    -ff  --fluentd_conf_folder=[enum]     Fluentd configuration folder (default: \"\")
"
  exit 0
}

auto_show_configuration() {

echo "
Note: get more deployment options with -h

Configuration (Automatic execution):
-=-=-=-=-=-=-=-=-=-=-=-=-=--=-=-=-=-
Automate fluentd against conf folder --> $fluentd_conf_folder
"
}

initial_deploy() {
      echo "

      ======>>> Performing initial benchmark deploy
      -==--=-==--=-=-==--=-==-=--==--=-=-=-=-=-=-=-==--=-=-=-=

      "

  export stress_profile="heavy"
  # export stress_profile="very-heavy" <-- optional for heavier stress
  export evacuate_node="false"
  export fluentd_image="docker.io/cognetive/origin-logging-fluentd:0.1"
  export fluent_conf_file="conf/collector/fluentd/fluentd.conf"
  export gologfilewatcher_image="docker.io/cognetive/go-log-file-watcher-with-symlink-support-v0"
  export fluentbit_image="fluent/fluent-bit:1.7-debug"
  export collector="fluentd"
  deploy
}

deploy_fluentd_with_conf() {
    fluent_conf_file=$1
    deploy_log_collector_fluentd "$fluentd_image" "$fluent_conf_file"

    # force containers redeploy
    oc scale --replicas=0 deployment fluentd
    oc scale --replicas=1 deployment fluentd
    oc scale --replicas=0 deployment capturestatistics
    oc scale --replicas=1 deployment capturestatistics

    # wait for enough results

    echo -e "===> Waiting for enough results"
    while : ; do
      CAPTURE_STATISTICS_POD=$(oc get pod -l app=capturestatistics -o jsonpath="{.items[0].metadata.name}")
      POD_READY=$(oc get pod "$CAPTURE_STATISTICS_POD" | grep Running)
      LOG_RESULTS=""
      RESULTS_COUNT=0
      if [ -n "$POD_READY" ]; then
        LOG_RESULTS=$(oc logs "$CAPTURE_STATISTICS_POD" | grep -A 6 "Report at");
        RESULTS_COUNT=$(($(echo -n "$LOG_RESULTS" | grep -c "Report at")))
        if (( RESULTS_COUNT > 3 )); then
          echo "==> we have $RESULTS_COUNT for running with $fluent_conf_file"
          echo "==> results are::"
          echo "$LOG_RESULTS"
          break;
        fi
      fi
      echo "we have $RESULTS_COUNT results - still waiting "
      sleep 30
    done
}

deploy_fluentd_partial() {

  # Initial benchmark deployment
  initial_deploy

  # re-deploy with each one of the configurations under `conf/collector/fluentd/partial/*`
  for fluent_conf_file in "$fluentd_conf_folder"/* ; do
    echo "

    ======>>> Deploying with configuration $fluent_conf_file
    -==--=-==--=-=-==--=-==-=--==--=-=-=-=-=-=-=-==--=-=-=-=

    "
    deploy_fluentd_with_conf "$fluent_conf_file"
  done
}


auto_deploy() {
  auto_show_configuration
  if [ -n "$fluentd_conf_folder" ]; then
    deploy_fluentd_partial;
  fi
}

AUTO_RUNNING="$(basename $(echo "$0" | sed 's/-//g'))"
if [[ "$AUTO_RUNNING" == "auto_execution.sh" ]]
then

  source ./contrib/deploy_functions.sh
  source ./contrib/expose_metrics.sh
  source ./deploy_to_openshift.sh

  #default parameters
  fluentd_conf_folder=""
  for i in "$@"
  do
  case $i in
      -ff=*|--fluentd_conf_folder=*) fluentd_conf_folder="${i#*=}"; shift ;;
      -h|--help|*) auto_show_usage ;;
  esac
  done

  auto_deploy "$@"
fi
