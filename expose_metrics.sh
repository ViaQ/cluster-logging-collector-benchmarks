oc project logstress
oc create -f config-map.yaml
oc apply -f deploy-go-app-service.yaml
oc apply -f service-monitor-gowatcher.yaml
oc apply -f deploy-fluentd-app-service.yaml
oc apply -f service-monitor-fluentd.yaml