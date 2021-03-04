package operator

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	log "github.com/go-logr/logr"
	harmonycloudcnv1alpha1 "harmonycloud.cn/nacos-operator/api/v1alpha1"
	myErrors "harmonycloud.cn/nacos-operator/pkg/errors"
	"harmonycloud.cn/nacos-operator/pkg/service/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IStatusClient interface {
}

type StatusClient struct {
	logger log.Logger
	client client.Client
}

func NewStatusClient(logger log.Logger, k8sService k8s.Services, client client.Client) *StatusClient {
	return &StatusClient{
		client: client,
		logger: logger,
	}
}

// 更新状态
func (c *StatusClient) UpdateStatus(nacos *harmonycloudcnv1alpha1.Nacos) {
	nacos.Status.Phase = harmonycloudcnv1alpha1.PhaseRunning
	// TODO
	myErrors.EnsureNormal(c.client.Status().Update(context.TODO(), nacos))

}

func (c *StatusClient) UpdateExceptionStatus(nacos *harmonycloudcnv1alpha1.Nacos, err *myErrors.Err) {

	var event harmonycloudcnv1alpha1.Event
	if len(nacos.Status.Conditions) == 0 {
		event = harmonycloudcnv1alpha1.Event{}
		nacos.Status.Event = append(nacos.Status.Event, event)
	} else {
		event = nacos.Status.Event[len(nacos.Status.Event)-1]
	}
	if event.Code == err.Code {
		event.LastTransitionTime.Time = time.Now()
	} else {
		event = harmonycloudcnv1alpha1.Event{
			Code:    err.Code,
			Status:  "false",
			Message: err.Msg,
			LastTransitionTime: metav1.Time{
				Time: time.Now(),
			},
			FirstAppearTime: metav1.Time{
				Time: time.Now(),
			},
		}
	}

	nacos.Status.Event[len(nacos.Status.Event)-1] = event
	// 设置为异常状态
	nacos.Status.Phase = harmonycloudcnv1alpha1.PhaseFailed
	e := c.client.Status().Update(context.TODO(), nacos)
	if e != nil {
		c.logger.V(-1).Info(e.Error())
	}

}
