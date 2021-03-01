package operator

import (
	log "github.com/go-logr/logr"
	"harmonycloud.cn/nacos-operator/pkg/service/k8s"
)

type IHealClient interface {
}

type HealClient struct {
}

func NewHealClient(logger log.Logger, k8sService k8s.Services) *HealClient {
	return &HealClient{}
}
