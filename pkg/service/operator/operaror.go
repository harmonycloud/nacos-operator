package operator

import (
	log "github.com/go-logr/logr"
	harmonycloudcnv1alpha1 "harmonycloud.cn/nacos-operator/api/v1alpha1"
	myErrors "harmonycloud.cn/nacos-operator/pkg/errors"
	"harmonycloud.cn/nacos-operator/pkg/service/k8s"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IOperatorClient interface {
	IKindClient
	ICheckClient
	IHealClient
	IStatusClient
}

type OperatorClient struct {
	KindClient   *KindClient
	CheckClient  *CheckClient
	HealClient   *HealClient
	StatusClient *StatusClient
}

func NewOperatorClient(logger log.Logger, clientset *kubernetes.Clientset, s *runtime.Scheme, client client.Client) *OperatorClient {
	service := k8s.NewK8sService(clientset, logger)
	return &OperatorClient{
		// 资源客户端
		KindClient: NewKindClient(logger, service, s),
		// 检测客户端
		CheckClient: NewCheckClient(logger, service),
		// 状态客户端
		StatusClient: NewStatusClient(logger, service, client),
		// 维护客户端
		HealClient: NewHealClient(logger, service),
	}
}

func (c *OperatorClient) MakeEnsure(nacos *harmonycloudcnv1alpha1.Nacos) {
	// 验证CR字段
	c.KindClient.ValidationField(nacos)

	switch nacos.Spec.Type {
	case TYPE_STAND_ALONE:
		c.KindClient.EnsureConfigmap(nacos)
		c.KindClient.EnsureStatefulset(nacos)
		c.KindClient.EnsureService(nacos)
	case TYPE_CLUSTER:
		c.KindClient.EnsureConfigmap(nacos)
		c.KindClient.EnsureStatefulsetCluster(nacos)
		c.KindClient.EnsureHeadlessServiceCluster(nacos)
		c.KindClient.EnsureClientService(nacos)
	default:
		panic(myErrors.New(myErrors.CODE_PARAMETER_ERROR, myErrors.MSG_PARAMETER_ERROT, "nacos.Spec.Type", nacos.Spec.Type))
	}
}

func (c *OperatorClient) PreCheck(nacos *harmonycloudcnv1alpha1.Nacos) {
	switch nacos.Status.Phase {
	case harmonycloudcnv1alpha1.PhaseFailed:
		// 失败，需要修复
		c.HealClient.MakeHeal(nacos)
	case harmonycloudcnv1alpha1.PhaseNone:
		// 初始化
		nacos.Status.Phase = harmonycloudcnv1alpha1.PhaseCreating
		panic(myErrors.New(myErrors.CODE_NORMAL, ""))
	case harmonycloudcnv1alpha1.PhaseScale:
	default:
		// TODO
	}
}

func (c *OperatorClient) CheckAndMakeHeal(nacos *harmonycloudcnv1alpha1.Nacos) {

	// 检查kind
	pods := c.CheckClient.CheckKind(nacos)
	// 检查nacos
	c.CheckClient.CheckNacos(nacos, pods)
}

func (c *OperatorClient) UpdateStatus(nacos *harmonycloudcnv1alpha1.Nacos) {
	c.StatusClient.UpdateStatusRunning(nacos)
}
