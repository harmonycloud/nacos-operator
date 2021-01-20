package operator

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	myErrors "harmonycloud.cn/nacos-operator/pkg/errors"

	log "github.com/go-logr/logr"
	harmonycloudcnv1alpha1 "harmonycloud.cn/nacos-operator/api/v1alpha1"
	"harmonycloud.cn/nacos-operator/pkg/service/k8s"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

const TYPE_STAND_ALONE = "standalone"
const TYPE_CLUSTER = "cluster"
const NACOS_PORT = 8848

type ClientEnsure interface {
	Ensure(nacos harmonycloudcnv1alpha1.Nacos)
	EnsureStatefulset(nacos harmonycloudcnv1alpha1.Nacos)
	EnsureConfigmap(nacos harmonycloudcnv1alpha1.Nacos)
}

type ClientEnsurer struct {
	k8sService k8s.Services
	logger     log.Logger
}

func NewClientEnsurer(logger log.Logger, k8sService k8s.Services) *ClientEnsurer {
	return &ClientEnsurer{
		k8sService: k8sService,
		logger:     logger,
	}
}

func (e *ClientEnsurer) Ensure(nacos *harmonycloudcnv1alpha1.Nacos) {
	switch nacos.Spec.Type {
	case TYPE_STAND_ALONE:
		e.EnsureConfigmap(nacos)
		e.EnsureStatefulset(nacos)
	case TYPE_CLUSTER:
		e.EnsureClusterConfigmap(nacos)
		e.EnsureStatefulset(nacos)
	}
}

func (e *ClientEnsurer) EnsureStatefulset(nacos *harmonycloudcnv1alpha1.Nacos) {
	ss := e.buildStatefulset(nacos)
	myErrors.EnsureNormal(e.k8sService.CreateOrUpdateStatefulSet(nacos.Namespace, ss))
}

func (e *ClientEnsurer) EnsureService(nacos *harmonycloudcnv1alpha1.Nacos) {
	ss := e.buildService(nacos)
	myErrors.EnsureNormal(e.k8sService.CreateOrUpdateService(nacos.Namespace, ss))
}

func (e *ClientEnsurer) EnsureConfigmap(nacos *harmonycloudcnv1alpha1.Nacos) {
	cm := e.buildConfigMap(nacos)
	myErrors.EnsureNormal(e.k8sService.CreateOrUpdateConfigMap(nacos.Namespace, cm))
}

func (e *ClientEnsurer) EnsureClusterConfigmap(nacos *harmonycloudcnv1alpha1.Nacos) {
	cm := e.buildConfigMap(nacos)
	myErrors.EnsureNormal(e.k8sService.CreateOrUpdateConfigMap(nacos.Namespace, cm))
}

func (e *ClientEnsurer) buildService(nacos *harmonycloudcnv1alpha1.Nacos) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        nacos.Name,
			Namespace:   nacos.Namespace,
			Labels:      nacos.Labels,
			Annotations: nacos.Annotations,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: nacos.Name,
					Port: NACOS_PORT,
				},
			},
		},
	}
	return svc
}
func (e *ClientEnsurer) buildStatefulset(nacos *harmonycloudcnv1alpha1.Nacos) *appv1.StatefulSet {
	MatchLabels := map[string]string{}
	MatchLabels["app"] = nacos.Name
	ss := appv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        nacos.Name,
			Namespace:   nacos.Namespace,
			Labels:      nacos.Labels,
			Annotations: nacos.Annotations,
		},
		Spec: appv1.StatefulSetSpec{
			Replicas: nacos.Spec.Replicas,
			Selector: &metav1.LabelSelector{MatchLabels: MatchLabels},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: MatchLabels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  nacos.Name,
							Image: nacos.Spec.Image,
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: NACOS_PORT,
								},
							},
						},
					},
				},
			},
		},
	}
	return &ss
}

func (e *ClientEnsurer) buildConfigMap(nacos *harmonycloudcnv1alpha1.Nacos) *v1.ConfigMap {
	data := make(map[string]string)

	data["application.properties"] = `server.servlet.contextPath=/nacos\n
server.port=8848\n`
	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        nacos.Name,
			Namespace:   nacos.Namespace,
			Labels:      nacos.Labels,
			Annotations: nacos.Annotations,
		},
		Data: data,
	}
	return &cm
}

func (e *ClientEnsurer) buildClusterConfigMap(nacos *harmonycloudcnv1alpha1.Nacos) *v1.ConfigMap {
	data := make(map[string]string)
	conf := ""
	for i := 0; i < int(*nacos.Spec.Replicas); i++ {
		conf = conf + fmt.Sprintf("%s-%d.%s:%d\n", nacos.Name, i, nacos.Name, nacos, 8848)
	}
	data["cluster.conf"] = conf
	cm := v1.ConfigMap{
		ObjectMeta: nacos.ObjectMeta,
		Data:       data,
	}
	return &cm
}
