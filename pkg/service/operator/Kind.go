package operator

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

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
const NACOS = "nacos"
const NACOS_PORT = 8848

type IKindClient interface {
	Ensure(nacos harmonycloudcnv1alpha1.Nacos)
	EnsureStatefulset(nacos harmonycloudcnv1alpha1.Nacos)
	EnsureConfigmap(nacos harmonycloudcnv1alpha1.Nacos)
}

type KindClient struct {
	k8sService k8s.Services
	logger     log.Logger
	scheme     *runtime.Scheme
}

func NewKindClient(logger log.Logger, k8sService k8s.Services, scheme *runtime.Scheme) *KindClient {
	return &KindClient{
		k8sService: k8sService,
		logger:     logger,
		scheme:     scheme,
	}
}

func (e *KindClient) generateLabels(name string, component string) map[string]string {
	return map[string]string{
		"app":        name,
		"middleware": NACOS,
		"component":  component,
	}
}

// 合并cr中的label 和 固定的label
func (e *KindClient) MergeLabels(allLabels ...map[string]string) map[string]string {
	res := map[string]string{}

	for _, labels := range allLabels {
		if labels != nil {
			for k, v := range labels {
				res[k] = v
			}
		}
	}
	return res
}

func (e *KindClient) generateName(nacos *harmonycloudcnv1alpha1.Nacos) string {
	return nacos.Name
}

func (e *KindClient) generateHeadlessServiceName(nacos *harmonycloudcnv1alpha1.Nacos) string {
	return nacos.Name
}

func (e *KindClient) EnsureStatefulsetCluster(nacos *harmonycloudcnv1alpha1.Nacos) {
	ss := e.buildStatefulset(nacos)
	ss = e.buildStatefulsetCluster(nacos, ss)
	myErrors.EnsureNormal(e.k8sService.CreateOrUpdateStatefulSet(nacos.Namespace, ss))
}

func (e *KindClient) EnsureStatefulset(nacos *harmonycloudcnv1alpha1.Nacos) {
	ss := e.buildStatefulset(nacos)
	myErrors.EnsureNormal(e.k8sService.CreateOrUpdateStatefulSet(nacos.Namespace, ss))
}

func (e *KindClient) EnsureService(nacos *harmonycloudcnv1alpha1.Nacos) {
	ss := e.buildService(nacos)
	myErrors.EnsureNormal(e.k8sService.CreateOrUpdateService(nacos.Namespace, ss))
}

func (e *KindClient) EnsureServiceCluster(nacos *harmonycloudcnv1alpha1.Nacos) {
	ss := e.buildService(nacos)
	myErrors.EnsureNormal(e.k8sService.CreateOrUpdateService(nacos.Namespace, ss))
}

func (e *KindClient) EnsureHeadlessServiceCluster(nacos *harmonycloudcnv1alpha1.Nacos) {
	ss := e.buildService(nacos)
	ss = e.buildHeadlessServiceCluster(ss)
	myErrors.EnsureNormal(e.k8sService.CreateOrUpdateService(nacos.Namespace, ss))
}

func (e *KindClient) EnsureConfigmap(nacos *harmonycloudcnv1alpha1.Nacos) {
	cm := e.buildConfigMap(nacos)
	myErrors.EnsureNormal(e.k8sService.CreateOrUpdateConfigMap(nacos.Namespace, cm))
}

func (e *KindClient) buildService(nacos *harmonycloudcnv1alpha1.Nacos) *v1.Service {
	labels := e.generateLabels(nacos.Name, NACOS)
	labels = e.MergeLabels(nacos.Labels, labels)
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        e.generateName(nacos),
			Namespace:   nacos.Namespace,
			Labels:      labels,
			Annotations: nacos.Annotations,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: nacos.Name,
					Port: NACOS_PORT,
				},
			},
			Selector: labels,
		},
	}
	controllerutil.SetControllerReference(nacos, svc, e.scheme)
	return svc
}
func (e *KindClient) buildStatefulset(nacos *harmonycloudcnv1alpha1.Nacos) *appv1.StatefulSet {
	// 生成label
	labels := e.generateLabels(nacos.Name, NACOS)
	// 合并cr中原有的label
	labels = e.MergeLabels(nacos.Labels, labels)

	var ss = &appv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        e.generateName(nacos),
			Namespace:   nacos.Namespace,
			Labels:      labels,
			Annotations: nacos.Annotations,
		},
		Spec: appv1.StatefulSetSpec{
			Replicas: nacos.Spec.Replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{{
						Name: "config",
						VolumeSource: v1.VolumeSource{
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{nacos.Name},
								Items: []v1.KeyToPath{
									{
										Key:  "application.properties",
										Path: "application.properties",
									},
								},
							},
						},
					},
					},
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
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "config",
									MountPath: "/home/nacos/conf/application.properties",
									SubPath:   "application.properties",
								},
							},
						},
					},
				},
			},
		},
	}

	controllerutil.SetControllerReference(nacos, ss, e.scheme)
	return ss
}

func (e *KindClient) buildConfigMap(nacos *harmonycloudcnv1alpha1.Nacos) *v1.ConfigMap {
	labels := e.generateLabels(nacos.Name, NACOS)
	labels = e.MergeLabels(nacos.Labels, labels)
	data := make(map[string]string)

	data["application.properties"] = nacos.Spec.Config
	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        e.generateName(nacos),
			Namespace:   nacos.Namespace,
			Labels:      labels,
			Annotations: nacos.Annotations,
		},
		Data: data,
	}
	controllerutil.SetControllerReference(nacos, &cm, e.scheme)
	return &cm
}

func (e *KindClient) buildStatefulsetCluster(nacos *harmonycloudcnv1alpha1.Nacos, ss *appv1.StatefulSet) *appv1.StatefulSet {
	ss.Spec.ServiceName = nacos.Name
	serivce := ""
	for i := 0; i < int(*nacos.Spec.Replicas); i++ {
		serivce = fmt.Sprintf("%v%v-%d.%v:%v ", serivce, e.generateName(nacos), i, e.generateName(nacos), NACOS_PORT)
	}
	env := []v1.EnvVar{
		{
			Name:  "NACOS_SERVERS",
			Value: serivce,
		},
	}
	ss.Spec.Template.Spec.Containers[0].Env = env
	return ss
}

func (e *KindClient) buildHeadlessServiceCluster(svc *v1.Service) *v1.Service {
	svc.Spec.ClusterIP = "None"
	return svc
}
