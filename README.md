# nacos-operator


## 快速开始
```bigquery
# 安装crd
make install

# 运行operator
make run

# 默认standalone+本地数据库+无存储，容器重启后数据丢失，请勿在生产环境部署
# 还支持以本地数据库运行、带存储、mysql、集群等配置，参考config/samples
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