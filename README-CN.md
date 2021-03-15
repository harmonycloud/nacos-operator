# nacos-operator

nacos-operator项目，快速在K8s上面部署构建nacos。

## 与nacos-k8s的项目区别
### 优点
- 通过operator快速构建nacos集群，指定简单的cr.yaml文件，既可以实现各种类型的nacos集群(数据库选型、standalone/cluster模式等)
- 增加一定的运维能力，在status中增加对nacos集群状态的检查、自动化运维等(后续扩展更多功能)

## 快速开始
```
# 直接使用helm方式安装operator
cd chart/nacos-operator && helm install nacos-operator . && cd ../..

# 如果没有helm, 使用kubectl进行安装, 默认安装在default下面
kubectl apply -f chart/nacos-operator/nacos-operator-all.yaml
```

### 启动单实例，standalone模式
查看cr文件
```
cat config/samples/nacos.yaml
apiVersion: harmonycloud.cn/v1alpha1
kind: Nacos
metadata:
  name: nacos
spec:
  type: standalone
  image: nacos/nacos-server:1.4.1
  replicas: 1

# 安装demo standalone模式
kubectl apply -f config/samples/nacos.yaml
```
查看nacos实例
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
清除
```
make demo clear=true
```
### 启动集群模式
```
cat config/samples/nacos_cluster.yaml

apiVersion: harmonycloud.cn/v1alpha1
kind: Nacos
metadata:
  name: nacos
spec:
  type: cluster
  image: nacos/nacos-server:1.4.1
  replicas: 3
```
```
# 创建nacos集群
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
  type: standalone
  image: nacos/nacos-server:1.4.1
  replicas: 1
  database:
    type: embedded
  # 启动数据卷，不然重启后数据丢失
  volume:
    enabled: true
    requests:
      storage: 1Gi
    storageClass: default
```

mysql数据库
```
apiVersion: harmonycloud.cn/v1alpha1
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
### 自定义配置
1. 通过环境变量配置 兼容nacos-docker项目， https://github.com/nacos-group/nacos-docker
   
    ```
    apiVersion: harmonycloud.cn/v1alpha1
    kind: Nacos
    metadata:
      name: nacos
    spec:
      type: standalone
      env:
      - key: JVM_XMS
        value: 2g
    ```

2. 通过properties文件配置

   https://github.com/nacos-group/nacos-docker/blob/master/build/bin/docker-startup.sh
   
   ```
   export CUSTOM_SEARCH_NAMES="application,custom"
   export CUSTOM_SEARCH_LOCATIONS=${BASE_DIR}/init.d/,file:${BASE_DIR}/conf/
   ```

    支持自定义配置文件，spec.config 会直接映射成custom.properties文件

    ```
    apiVersion: harmonycloud.cn/v1alpha1
    kind: Nacos
    metadata:
      name: nacos
    spec:
    ...
      config:|
        management.endpoints.web.exposure.include=*
    ```

## 开发文档
```
# 安装crd
make install

# 以源码方式运行operator
# make run
```


## FAQ
1. 设置readiness和liveiness集群出问题

    最后一个实例无法ready，搜索了下issus，发现需要以下设置
    ```
    nacos.naming.data.warmup=false
    ```
    
    设置了以后发现，pod能够running，但是集群状态始终无法同步，不同节点出现不同leader；所以暂时不开启readiness和liveiness
   

2. 组集群失败
```
java.lang.IllegalStateException: unable to find local peer: nacos-1.nacos-headless.shenkonghui.svc.cluster.local:8848, all peers: []```