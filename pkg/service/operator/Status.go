package operator

import (
	log "github.com/go-logr/logr"
	"harmonycloud.cn/nacos-operator/pkg/service/k8s"
)

type IStatusClient interface {
}

type StatusClient struct {
}

func NewStatusClient(logger log.Logger, k8sService k8s.Services) *StatusClient {
	return &StatusClient{}
}
