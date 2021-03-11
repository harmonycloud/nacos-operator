# nacos-operator

nacos-operator, quickly deploy and build nacos on K8s.

[中文文档](./README-CN.md)
## Quick start
```
# install crd
make install

# run operator
make run
```
### Start single instance, standalone mode
```
# View cr file
cat config/samples/harmonycloud.cn_v1alpha1_nacos.yaml
apiVersion: harmonycloud.cn/v1alpha1
kind: Nacos
metadata:
  name: nacos
spec:
  type: standalone
  image: nacos/nacos-server:1.4.1
  replicas: 1
  # enable the local database
  enableEmbedded: true

# Install demo standalone mode
make demo type=standalone

# View nacos instance
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
  version: 1.4.0
```
clear
```
make demo clear=true
```
### Start cluster mode
```
cat config/samples/harmonycloud.cn_v1alpha1_nacos_cluster.yaml

apiVersion: harmonycloud.cn/v1alpha1
kind: Nacos
metadata:
  name: nacos
spec:
  type: cluster
  image: nacos/nacos-server:1.4.1
  replicas: 3
  enableEmbedded: true
  
# Create nacos cluster
make demo type=cluster


kubectl get po -o wide

NAME             READY   STATUS    RESTARTS   AGE    IP               NODE         NOMINATED NODE   READINESS GATES
nacos-0          1/1     Running   0          111s   10.168.247.39    slave-100    <none>           <none>
nacos-1          1/1     Running   0          109s   10.168.152.186   master-212   <none>           <none>
nacos-2          1/1     Running   0          108s   10.168.207.209   slave-214    <none>           <none>

# Wait for the cluster components, check the cr details again
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

clear
```
make demo clear=true
```
## Configuration
### Set mode
Currently supports standalone and cluster modes

By configuring spec.type as standalone/cluster

### Database configuration
embedded database
```
apiVersion: harmonycloud.cn/v1alpha1
kind: Nacos
metadata:
  name: nacos
spec:
  ...
  enableEmbedded: true
  # If you need to persist data, you need to configure volumeClaimTemplates, otherwise the pod restarts and the data will be lost
  volumeClaimTemplates:
  - metadata:
      name: db
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "default"
      resources:
        requests:
          storage: 100Mi
```

mysql database
```
apiVersion: harmonycloud.cn/v1alpha1
kind: Nacos
metadata:
  name: nacos
spec:
...
  config: |
    spring.datasource.platform=mysql
    db.num=1
    db.url.0=jdbc:mysql://mysql:3306/nacos?characterEncoding=utf8&connectTimeout=1000&socketTimeout=3000&autoReconnect=true
    db.user=root
    db.password=123456
```
### Custom configuration
Support custom configuration file, spec.config will be directly mapped to application.properties file

## FAQ
1. Problem setting readiness and liveiness cluster

   The last instance cannot be ready. I searched for issus and found that the following settings are required
    ```
    nacos.naming.data.warmup=false
    ```

   After setting it up, it is found that the pod can run, but the cluster status cannot always be synchronized, and different nodes have different leaders; therefore, readiness and liveiness are not enabled for the time being