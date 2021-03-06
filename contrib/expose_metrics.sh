#!/bin/bash

# expose metrics to OCP prometheus
expose_metrics_to_prometheus() {
  oc project logstress
  oc create -f conf/expose_metrics/config-map.yaml
  oc apply -f conf/expose_metrics/deploy-go-app-service.yaml
  oc apply -f conf/expose_metrics/service-monitor-gowatcher.yaml
  oc apply -f conf/expose_metrics/deploy-fluentd-app-service.yaml
  oc apply -f conf/expose_metrics/service-monitor-fluentd.yaml
}
