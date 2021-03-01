# nacos-operator


## 快速开始
```bigquery
# 安装crd
make install

# 运行operator
make run

# 打开config/samples/harmonycloud.cn_v1alpha1_nacos.yaml 配置spec.config,设置mysql信息，并导入sql
# 运行demo
make demo

$ kubectl get pod  -o wide
NAME                 READY   STATUS    RESTARTS   AGE    IP               NODE        NOMINATED NODE   READINESS GATES
nacos-standalone-0   1/1     Running   0          84s    10.168.247.38    slave-100   <none>           <none>
```

## makefile说明
### sdk框架自带
- make generate #生成代码
- make install  #安装crd
- make docker-build #打包docker
- make docker-push #推送docker
- make deploy #部署全部(在xxx-operator-system下面)


### 自定义实现
- make demo #创建demo standalone 模式
- make demo-cluster #创建demo cluster模式