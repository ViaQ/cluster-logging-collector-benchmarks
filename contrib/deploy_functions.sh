#!/bin/bash

# Selecting worker node to use
select_node_to_use() {
  echo "--> Selecting node to use" 
  NODE_TO_USE=$(oc get nodes --selector='!node-role.kubernetes.io/master' --sort-by=".metadata.name" -o=jsonpath='{.items[0].metadata.name}')
  echo "Using node: $NODE_TO_USE" 
}

# configure containers log-max-size to 10MB
configure_workers_log_rotation() {
  echo "--> Configuring log-max-size for nodes to $1"
  oc label mcp worker custom-crio=true --overwrite
  oc delete ctrcfg logsizemax
  oc create -f - <<EOF
apiVersion: machineconfiguration.openshift.io/v1
kind: ContainerRuntimeConfig
metadata:
  name: logsizemax
spec:
  machineConfigPoolSelector:
    matchLabels:
      custom-crio: "true"
  containerRuntimeConfig:
    logSizeMax: "$1"
EOF
  oc get ctrcfg
}

# delete logstress project if it exists (to get new fresh deployment)
delete_logstress_project_if_exists() {
PROJECT_LOG_STRESS=$(oc get project | grep logstress)
if [ -n "$PROJECT_LOG_STRESS" ]; then
  echo "--> Deleting log stress namespace"
  oc delete project logstress
  while : ; do
    PROJECT_LOG_STRESS=$(oc get project | grep logstress)
    if [ -z "$PROJECT_LOG_STRESS" ]; then break; fi
    sleep 1
  done
fi
}

# create and switch context to logstress project
create_logstress_project() {
  echo "--> Creating log stress namespace"
  oc label nodes "$NODE_TO_USE" logstress=true --overwrite
  oc adm new-project --node-selector='logstress=true' logstress
  oc project logstress
}

# set credentials (allow privileged) 
set_credentials() {
  oc adm policy add-scc-to-user privileged -z default
  oc adm policy add-cluster-role-to-user cluster-reader -z default
}

# deploy log stress containers
deploy_logstress() {
  DEPLOY_YAML=conf/stressor/logstress-template.yaml

  echo "--> Deploying $DEPLOY_YAML - with ($1 $2 $3 $4)"
  rm -f log-stressor.zip
  rm -f log-stressor
  go build -ldflags "-s -w" go/log-stressor/log-stressor.go
  zip log-stressor.zip  log-stressor
  oc delete configmap --ignore-not-found=true log-stressor-binary-zip
  oc create configmap log-stressor-binary-zip --from-file=log-stressor.zip
  rm -f log-stressor.zip
  rm -f log-stressor

  oc process -f $DEPLOY_YAML \
    -p number_heavy_stress_containers="$1" \
    -p heavy_containers_msg_per_sec="$2" \
    -p number_low_stress_containers="$3" \
    -p low_containers_msg_per_sec="$4" \
    | oc apply -f -
}

# deploy log collector (fluentd) container
deploy_log_collector_fluentd() {
  DEPLOY_YAML=conf/collector/fluentd/fluentd-template.yaml

  echo "--> Deploying $DEPLOY_YAML - with ($1 $2)"
  oc delete configmap --ignore-not-found=true fluentd-config
  oc create configmap fluentd-config --from-file="$2"
  oc process -f $DEPLOY_YAML \
    -p fluentd_image="$1" \
    | oc apply -f -
}

#deploy gologfilewatch container
deploy_gologfilewatcher() {
	DEPLOY_YAML=conf/expose_metrics/gologfilewatcher-template.yaml

	echo "--> Deploying $DEPLOY_YAML -with ($1)"
	oc process -f $DEPLOY_YAML \
		-p gologfilewatcher_image="$1" \
		| oc apply -f -
}

# deploy log collector (fluentbit) container
deploy_log_collector_fluentbit() {
  DEPLOY_YAML=conf/collector/fluentbit/fluentbit-template.yaml

  echo "--> Deploying $DEPLOY_YAML - with ($1)"
  oc delete configmap --ignore-not-found=true fluentbit-config
  oc create configmap fluentbit-config --from-file="$2" --from-file=conf/collector/fluentbit/fluentbit.parsers.conf
  oc process -f $DEPLOY_YAML \
    -p fluentbit_image="$1" \
    | oc apply -f -
}

# deploy capture statistics container
deploy_capture_statistics() {
  DEPLOY_YAML=conf/monitor/capture-statistics-template.yaml

  echo "--> Deploying $DEPLOY_YAML - with ($1)"
  rm -f check-logs-sequence.zip
  rm -f check-logs-sequence
  go get -u github.com/papertrail/go-tail
  go build -ldflags "-s -w" go/check-logs-sequence/check-logs-sequence.go
  zip check-logs-sequence.zip  check-logs-sequence
  oc delete configmap --ignore-not-found=true check-logs-sequence-binary-zip
  oc create configmap check-logs-sequence-binary-zip --from-file=check-logs-sequence.zip
  rm -f check-logs-sequence.zip
  rm -f check-logs-sequence
  oc process -f $DEPLOY_YAML \
  -p number_of_log_lines_between_reports="$1" \
  | oc apply -f -
}

evacuate_node_for_performance_tests() {
  echo "--> Evacuating $NODE_TO_USE"
  oc get pods --all-namespaces -o wide | grep "$NODE_TO_USE"
  
  oc adm cordon "$NODE_TO_USE"
  oc adm drain "$NODE_TO_USE" --pod-selector='app notin (low-log-stress,heavy-log-stress,fluentd,capturestatistics)' --ignore-daemonsets=true --delete-local-data --force
}

return_node_to_normal() {
  echo "--> Allow scheduling on $NODE_TO_USE"
  oc adm uncordon "$NODE_TO_USE"
  while : ; do
    NODE_SCHEDULING_DISABLED=$(oc get nodes --selector='!node-role.kubernetes.io/master' | grep SchedulingDisabled)
    if [ -z "$NODE_SCHEDULING_DISABLED" ]; then break; fi
    sleep 1
  done
  oc get nodes --selector='!node-role.kubernetes.io/master'
}


# print pod status
print_pods_status () {
  echo -e "\n"
  oc get pods
}

# print usage instructions
print_usage_instructions () {
  CAPTURE_STATISTICS_POD=$(oc get pod -l app=capturestatistics -o jsonpath="{.items[0].metadata.name}")

  echo -e "\n\nExplore logs of relevant pods ^^^"
  echo -e "Waiting for $CAPTURE_STATISTICS_POD to become ready"
  while : ; do
    POD_READY=$(oc get pod "$CAPTURE_STATISTICS_POD" | grep Running)
    if [ -n "$POD_READY" ]; then break; fi
    sleep 1
  done
  echo -e "Tailing pod $CAPTURE_STATISTICS_POD using command:"
  command="oc logs -f $CAPTURE_STATISTICS_POD"
  echo -e "$command"
  $command
}
