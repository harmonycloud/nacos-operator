# nacos-operator

## makefile说明
### 框架自带
- make generate #生成代码
- make install  #安装crd
- make docker-build #打包docker
- make docker-push #推送docker
- make deploy #部署全部(在xxx-operator-system下面)

  
### 自定义实现
- make demo #创建demo
- make rdemo #删除demo
- make image 