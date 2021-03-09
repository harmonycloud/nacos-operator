# nacos-operator

nacos-operator，快速在K8s上面部署构建nacos

## 快速开始
```
# 安装crd
make install

# 运行operator
make run
```
### 启动单实例，standalone模式
```
# 查看cr文件
cat config/samples/harmonycloud.cn_v1alpha1_nacos.yaml
apiVersion: harmonycloud.cn/v1alpha1
kind: Nacos
metadata:
  name: nacos
spec:
  type: standalone
  image: nacos/nacos-server:1.4.1
  replicas: 1
  # 开启本地数据库
  enableEmbedded: true

# 安装demo standalone模式
make demo type=standalone

# 查看nacos实例
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
清除
```
make demo clear=true
```
### 启动集群模式
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
  
# 创建nacos集群
make demo type=cluster


kubectl get po -o wide

NAME             READY   STATUS    RESTARTS   AGE    IP               NODE         NOMINATED NODE   READINESS GATES
nacos-0          1/1     Running   0          111s   10.168.247.39    slave-100    <none>           <none>
nacos-1          1/1     Running   0          109s   10.168.152.186   master-212   <none>           <none>
nacos-2          1/1     Running   0          108s   10.168.207.209   slave-214    <none>           <none>

# 等待集群组件，再次cr详情
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

清除
```
make demo clear=true
```
## 配置
### 设置模式
目前支持standalone和cluster模式

通过配置spec.type 为 standalone/cluster

### 数据库配置
embedded数据库
```
apiVersion: harmonycloud.cn/v1alpha1
kind: Nacos
metadata:
  name: nacos
spec:
  ...
  enableEmbedded: true
  # 如果需要持久化数据，需要配置volumeClaimTemplates，否则pod重启数据丢失
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

mysql数据库
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
### 自定义配置
支持自定义配置文件，spec.config 会直接映射成application.properties文件

## FAQ
1. 设置readiness和liveiness集群出问题

    最后一个实例无法ready，搜索了下issus，发现需要以下设置
    ```
    nacos.naming.data.warmup=false
    ```
    
    设置了以后发现，pod能够running，但是集群状态始终无法同步，不同节点出现不同leader；所以暂时不开启readiness和liveiness