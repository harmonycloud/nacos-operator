package operator

import (
	log "github.com/go-logr/logr"
	harmonycloudcnv1alpha1 "harmonycloud.cn/nacos-operator/api/v1alpha1"
	myErrors "harmonycloud.cn/nacos-operator/pkg/errors"
	"harmonycloud.cn/nacos-operator/pkg/service/k8s"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

type IOperatorClient interface {
}

type OperatorClient struct {
	KindClient   *KindClient
	CheckClient  *CheckClient
	HealClient   *HealClient
	StatusClient *StatusClient
}

func NewOperatorClient(logger log.Logger, clientset *kubernetes.Clientset, s *runtime.Scheme) *OperatorClient {
	service := k8s.NewK8sService(clientset, logger)
	return &OperatorClient{
		KindClient:   NewKindClient(logger, service, s),
		StatusClient: NewStatusClient(logger, service),
	}
}

func (c *OperatorClient) MakeEnsure(nacos *harmonycloudcnv1alpha1.Nacos) {
	switch nacos.Spec.Type {
	case TYPE_STAND_ALONE:
		c.KindClient.EnsureConfigmap(nacos)
		c.KindClient.EnsureStatefulset(nacos)
		c.KindClient.EnsureService(nacos)
	case TYPE_CLUSTER:
		c.KindClient.EnsureConfigmap(nacos)
		c.KindClient.EnsureStatefulsetCluster(nacos)
		//c.KindClient.EnsureServiceCluster(nacos)
		c.KindClient.EnsureHeadlessServiceCluster(nacos)
	default:
		panic(myErrors.New(myErrors.PARAMETER_ERROR, myErrors.MSG_PARAMETER_ERROT, "nacos.Spec.Type", nacos.Spec.Type))
	}
}

func (c *OperatorClient) CheckAndMakeHeal(nacos *harmonycloudcnv1alpha1.Nacos) {
}

func (c *OperatorClient) UpdateStatus(nacos *harmonycloudcnv1alpha1.Nacos) {
}
