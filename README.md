# nacos-operator


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
apiVersion: harmonycloud.cn.harmonycloud.cn/v1alpha1
kind: Nacos
metadata:
  name: nacos
spec:
  type: standalone
  image: 10.1.11.100/middleware/nacose:1.4.0
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

apiVersion: harmonycloud.cn.harmonycloud.cn/v1alpha1
kind: Nacos
metadata:
  name: nacos
spec:
  type: cluster
  image: 10.1.11.100/middleware/nacose:1.4.1
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


## FAQ

