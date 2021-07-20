# nacos-operator

nacos-operator, quickly deploy and build nacos on K8s.

[中文文档](./README-CN.md)
## Difference with nacos-k8s
### advantage
- Quickly build a nacos cluster through the operator, specify a simple cr.yaml file, and realize various types of nacos clusters (database selection, standalone/cluster mode, etc.)
- Add a certain amount of operation and maintenance capabilities, add the inspection of the nacos cluster status, automatic operation and maintenance, etc. in the status (more functions will be expanded later)

## Quick start
```
# Install operator directly using helm
cd chart/nacos-operator && helm install nacos-operator . && cd ../..

# If there is no helm, use kubectl to install it, and install it under default by default
kubectl apply -f chart/nacos-operator/nacos-operator-all.yaml
```

### Start single instance, standalone mode
View cr file
```
cat config/samples/nacos.yaml
apiVersion: nacos.io/v1alpha1
kind: Nacos
metadata:
  name: nacos
spec:
  type: standalone
  image: nacos/nacos-server:1.4.1
  replicas: 1

# Install demo standalone mode
kubectl apply -f config/samples/nacos.yaml
```

View nacos instance
```
kubectl get nacos
NAME    REPLICAS   READY     TYPE         DBTYPE   VERSION   CREATETIME
nacos   1          Running   standalone            1.4.1     2021-03-14T09:21:49Z

kubectl get pod  -o wide
NAME                 READY   STATUS    RESTARTS   AGE    IP               NODE        NOMINATED NODE   READINESS GATES
nacos-0   1/1     Running   0          84s    10.168.247.38    slave-100   <none>           <none>

kubectl get nacos nacos -o yaml
...
status
  conditions:
  - instance: 10.168.247.38
    nodeName: slave-100
    podName: nacos-0
    status: "true"
    type: leader
  phase: Running
  version: 1.4.1
```
Clear
```
make demo clear=true
```
### Start cluster mode
```
cat config/samples/nacos_cluster.yaml

apiVersion: nacos.io/v1alpha1
kind: Nacos
metadata:
  name: nacos
spec:
  type: cluster
  image: nacos/nacos-server:1.4.1
  replicas: 3
```
```
# Create nacos cluster
kubectl apply -f config/samples/nacos_cluster.yaml

kubectl get po -o wide
NAME             READY   STATUS    RESTARTS   AGE    IP               NODE         NOMINATED NODE   READINESS GATES
nacos-0          1/1     Running   0          111s   10.168.247.39    slave-100    <none>           <none>
nacos-1          1/1     Running   0          109s   10.168.152.186   master-212   <none>           <none>
nacos-2          1/1     Running   0          108s   10.168.207.209   slave-214    <none>           <none>

kubectl get nacos
NAME    REPLICAS   READY     TYPE      DBTYPE   VERSION   CREATETIME
nacos   3          Running   cluster            1.4.1     2021-03-14T09:33:09Z

kubectl get nacos nacos -o yaml -w
...
status:
  conditions:
  - instance: 10.168.247.39
    nodeName: slave-100
    podName: nacos-0
    status: "true"
    type: leader
  - instance: 10.168.152.186
    nodeName: master-212
    podName: nacos-1
    status: "true"
    type: Followers
  - instance: 10.168.207.209
    nodeName: slave-214
    podName: nacos-2
    status: "true"
    type: Followers
  event:
  - code: -1
    firstAppearTime: "2021-03-05T08:35:03Z"
    lastTransitionTime: "2021-03-05T08:35:06Z"
    message: The number of ready pods is too small[]
    status: false
  - code: 200
    firstAppearTime: "2021-03-05T08:36:09Z"
    lastTransitionTime: "2021-03-05T08:36:48Z"
    status: true
  phase: Running
  version: 1.4.1
```

Clear
```
make demo clear=true
```
## Configuration
### Setting Mode
Currently supports standalone and cluster modes

By configuring spec.type as standalone/cluster

### Database configuration
embedded
```
apiVersion: nacos.io/v1alpha1
kind: Nacos
metadata:
  name: nacos
spec:
  type: standalone
  image: nacos/nacos-server:1.4.1
  replicas: 1
  database:
    type: embedded
  # Start the data volume, otherwise the data will be lost after restart
  volume:
    enabled: true
    requests:
      storage: 1Gi
    storageClass: default
```

mysql
```
apiVersion: nacos.io/v1alpha1
kind: Nacos
metadata:
  name: nacos
spec:
  type: standalone
  image: nacos/nacos-server:1.4.1
  replicas: 1
  database:
    type: mysql
    mysqlHost: mysql
    mysqlDb: nacos
    mysqlUser: root
    mysqlPort: "3306"
    mysqlPassword: "123456"
```
### Custom configuration
1. Configure through environment variables, compatible with nacos-docker project, https://github.com/nacos-group/nacos-docker

    ```
    apiVersion: nacos.io/v1alpha1
    kind: Nacos
    metadata:
      name: nacos
    spec:
      type: standalone
      env:
      - key: JVM_XMS
        value: 2g
    ```

2. Configure through the properties file

   https://github.com/nacos-group/nacos-docker/blob/master/build/bin/docker-startup.sh

   ```
   export CUSTOM_SEARCH_NAMES="application,custom"
   export CUSTOM_SEARCH_LOCATIONS=${BASE_DIR}/init.d/,file:${BASE_DIR}/conf/
   ```

   Support custom configuration file, spec.config will be directly mapped to custom.properties file

    ```
    apiVersion: nacos.io/v1alpha1
    kind: Nacos
    metadata:
      name: nacos
    spec:
    ...
      config:|
        management.endpoints.web.exposure.include=*
    ```

## Development Document
```
# Install crd
make install

# Run the operator as source code
# make run
```