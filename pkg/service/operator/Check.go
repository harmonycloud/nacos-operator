package operator

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"

	log "github.com/go-logr/logr"
	harmonycloudcnv1alpha1 "harmonycloud.cn/nacos-operator/api/v1alpha1"
	myErrors "harmonycloud.cn/nacos-operator/pkg/errors"
	"harmonycloud.cn/nacos-operator/pkg/service/k8s"
	nacosClient "harmonycloud.cn/nacos-operator/pkg/service/nacos"
)

type ICheckClient interface {
	Check(nacos *harmonycloudcnv1alpha1.Nacos)
}

type CheckClient struct {
	k8sService  k8s.Services
	logger      log.Logger
	nacosClient nacosClient.NacosClient
}

func NewCheckClient(logger log.Logger, k8sService k8s.Services) *CheckClient {
	return &CheckClient{
		k8sService: k8sService,
		logger:     logger,
	}
}

func (c CheckClient) CheckKind(nacos *harmonycloudcnv1alpha1.Nacos) []corev1.Pod {
	// ss数量和cr副本数匹配
	ss, err := c.k8sService.GetStatefulSet(nacos.Namespace, nacos.Name)
	myErrors.EnsureNormal(err)

	if ss.Spec.Replicas != nacos.Spec.Replicas {
		panic(myErrors.New(myErrors.CODE_ERR_UNKNOW, "cr replicas is not equal ss replicas"))

	}

	// 检查正常的pod数量，根据实际情况。如果单实例，必须要有1个;集群要1/2以上
	pods, err := c.k8sService.GetStatefulSetReadPod(nacos.Namespace, nacos.Name)
	if len(pods) < (int(*nacos.Spec.Replicas)+1)/2 {
		panic(myErrors.New(myErrors.CODE_ERR_UNKNOW, "The number of ready pods is too small"))
	} else if len(pods) != int(*nacos.Spec.Replicas) {
		c.logger.V(0).Info("pod num is not right")
	}
	return pods
}

func (c CheckClient) CheckNacos(pods []corev1.Pod) {
	// 检查nacos是否访问通
	for _, pod := range pods {
		str, err := c.nacosClient.GetClusterNodes(pod.Status.PodIP)
		myErrors.EnsureNormal(err)
		node := nacosClient.NacosClusterNodes{}
		myErrors.EnsureNormal(json.Unmarshal([]byte(str), &node))
		if node.Code != 200 {
			panic(myErrors.New(myErrors.CODE_ERR_UNKNOW, myErrors.MSG_NACOS_UNREACH, pod.Status.PodIP))
		}
	}
	//TODO
}
