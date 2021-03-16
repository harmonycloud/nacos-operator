package operator

import (
	"fmt"
	"strconv"

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
const RAFT_PORT = 7848

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

func (e *KindClient) generateAnnoation() map[string]string {
	return map[string]string{}
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

func (e *KindClient) generateHeadlessSvcName(nacos *harmonycloudcnv1alpha1.Nacos) string {
	return fmt.Sprintf("%s-headless", nacos.Name)
}

// CR格式验证
func (e *KindClient) ValidationField(nacos *harmonycloudcnv1alpha1.Nacos) {

	if nacos.Spec.Type == "" {
		nacos.Spec.Type = "standalone"
	}

	// 默认设置内置数据库
	if nacos.Spec.Database.TypeDatabase == "" {
		nacos.Spec.Database.TypeDatabase = "embedded"
	}
	// mysql设置默认值
	if nacos.Spec.Database.TypeDatabase == "mysql" {
		if nacos.Spec.Database.MysqlHost == "" {
			nacos.Spec.Database.MysqlHost = "127.0.0.1"
		}
		if nacos.Spec.Database.MysqlUser == "" {
			nacos.Spec.Database.MysqlUser = "root"
		}
		if nacos.Spec.Database.MysqlDb == "" {
			nacos.Spec.Database.MysqlDb = "nacos"
		}
		if nacos.Spec.Database.MysqlPassword == "" {
			nacos.Spec.Database.MysqlPassword = "123456"
		}
		if nacos.Spec.Database.MysqlPort == "" {
			nacos.Spec.Database.MysqlPort = "3306"
		}
	}
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
	myErrors.EnsureNormal(e.k8sService.CreateIfNotExistsService(nacos.Namespace, ss))
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
	if nacos.Spec.Config != "" {
		cm := e.buildConfigMap(nacos)
		myErrors.EnsureNormal(e.k8sService.CreateIfNotExistsConfigMap(nacos.Namespace, cm))
	}
}

func (e *KindClient) buildService(nacos *harmonycloudcnv1alpha1.Nacos) *v1.Service {
	labels := e.generateLabels(nacos.Name, NACOS)
	labels = e.MergeLabels(nacos.Labels, labels)

	annotations := e.MergeLabels(e.generateAnnoation(), nacos.Annotations)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        e.generateHeadlessSvcName(nacos),
			Namespace:   nacos.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1.ServiceSpec{
			PublishNotReadyAddresses: true,
			Ports: []v1.ServicePort{
				{
					Name:     "client",
					Port:     NACOS_PORT,
					Protocol: "TCP",
				},
				{
					Name:     "rpc",
					Port:     RAFT_PORT,
					Protocol: "TCP",
				},
			},
			Selector: labels,
		},
	}
	myErrors.EnsureNormal(controllerutil.SetControllerReference(nacos, svc, e.scheme))
	return svc
}
func (e *KindClient) buildStatefulset(nacos *harmonycloudcnv1alpha1.Nacos) *appv1.StatefulSet {
	// 生成label
	labels := e.generateLabels(nacos.Name, NACOS)
	// 合并cr中原有的label
	labels = e.MergeLabels(nacos.Labels, labels)

	// 设置默认的环境变量
	env := append(nacos.Spec.Env, v1.EnvVar{
		Name:  "PREFER_HOST_MODE",
		Value: "hostname",
	})

	// 数据库设置
	if nacos.Spec.Database.TypeDatabase == "embedded" {
		env = append(env, v1.EnvVar{
			Name:  "EMBEDDED_STORAGE",
			Value: "embedded",
		})
	} else if nacos.Spec.Database.TypeDatabase == "mysql" {
		env = append(env, v1.EnvVar{
			Name:  "MYSQL_SERVICE_HOST",
			Value: nacos.Spec.Database.MysqlHost,
		})

		env = append(env, v1.EnvVar{
			Name:  "MYSQL_SERVICE_PORT",
			Value: nacos.Spec.Database.MysqlPort,
		})

		env = append(env, v1.EnvVar{
			Name:  "MYSQL_SERVICE_DB_NAME",
			Value: nacos.Spec.Database.MysqlDb,
		})

		env = append(env, v1.EnvVar{
			Name:  "MYSQL_SERVICE_USER",
			Value: nacos.Spec.Database.MysqlUser,
		})

		env = append(env, v1.EnvVar{
			Name:  "MYSQL_SERVICE_PASSWORD",
			Value: nacos.Spec.Database.MysqlPassword,
		})
	}

	// 启动模式 ，默认cluster
	if nacos.Spec.Type == TYPE_STAND_ALONE {
		env = append(env, v1.EnvVar{
			Name:  "MODE",
			Value: "standalone",
		})
	} else {
		env = append(env, v1.EnvVar{
			Name:  "NACOS_REPLICAS",
			Value: strconv.Itoa(int(*nacos.Spec.Replicas)),
		})
	}

	var ss = &appv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        e.generateName(nacos),
			Namespace:   nacos.Namespace,
			Labels:      labels,
			Annotations: nacos.Annotations,
		},
		Spec: appv1.StatefulSetSpec{
			PodManagementPolicy: "Parallel",
			Replicas:            nacos.Spec.Replicas,
			Selector:            &metav1.LabelSelector{MatchLabels: labels},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{},
					Containers: []v1.Container{
						{
							Name:  nacos.Name,
							Image: nacos.Spec.Image,
							Ports: []v1.ContainerPort{
								{
									Name:          "client",
									ContainerPort: NACOS_PORT,
									Protocol:      "TCP",
								},
								{
									Name:          "rpc",
									ContainerPort: RAFT_PORT,
									Protocol:      "TCP",
								},
							},
							Env:            env,
							LivenessProbe:  nacos.Spec.LivenessProbe,
							ReadinessProbe: nacos.Spec.ReadinessProbe,
							VolumeMounts:   []v1.VolumeMount{},
						},
					},
				},
			},
		},
	}

	// 设置存储
	if nacos.Spec.Volume.Enabled {
		ss.Spec.VolumeClaimTemplates = append(ss.Spec.VolumeClaimTemplates, v1.PersistentVolumeClaim{
			Spec: v1.PersistentVolumeClaimSpec{
				//VolumeName:       "db",
				StorageClassName: nacos.Spec.Volume.StorageClass,
				AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
				Resources: v1.ResourceRequirements{
					Requests: nacos.Spec.Volume.Requests,
				},
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:   "db",
				Labels: labels,
			},
		})

		localVolum := v1.VolumeMount{
			Name:      "db",
			MountPath: "/home/nacos/data",
		}
		ss.Spec.Template.Spec.Containers[0].VolumeMounts = append(ss.Spec.Template.Spec.Containers[0].VolumeMounts, localVolum)
	}

	//probe := &v1.Probe{
	//	InitialDelaySeconds: 10,
	//	PeriodSeconds:       5,
	//	TimeoutSeconds:      4,
	//	FailureThreshold:    5,
	//	Handler: v1.Handler{
	//		HTTPGet: &v1.HTTPGetAction{
	//			Port: intstr.IntOrString{IntVal: NACOS_PORT},
	//			Path: "/nacos/actuator/health/",
	//		},
	//		//TCPSocket: &v1.TCPSocketAction{
	//		//	Port: intstr.IntOrString{IntVal: NACOS_PORT},
	//		//},
	//	},
	//}

	//if nacos.Spec.LivenessProbe == nil {
	//	ss.Spec.Template.Spec.Containers[0].LivenessProbe = probe
	//}
	//if nacos.Spec.ReadinessProbe == nil {
	//	ss.Spec.Template.Spec.Containers[0].ReadinessProbe = probe
	//}

	if nacos.Spec.Config != "" {
		ss.Spec.Template.Spec.Volumes = append(ss.Spec.Template.Spec.Volumes, v1.Volume{
			Name: "config",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{Name: nacos.Name},
					Items: []v1.KeyToPath{
						{
							Key:  "custom.properties",
							Path: "custom.properties",
						},
					},
				},
			},
		})
		ss.Spec.Template.Spec.Containers[0].VolumeMounts = append(ss.Spec.Template.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
			Name:      "config",
			MountPath: "/home/nacos/init.d/custom.properties",
			SubPath:   "custom.properties",
		})
	}
	myErrors.EnsureNormal(controllerutil.SetControllerReference(nacos, ss, e.scheme))
	return ss
}

func (e *KindClient) buildConfigMap(nacos *harmonycloudcnv1alpha1.Nacos) *v1.ConfigMap {
	labels := e.generateLabels(nacos.Name, NACOS)
	labels = e.MergeLabels(nacos.Labels, labels)
	data := make(map[string]string)

	data["custom.properties"] = nacos.Spec.Config

	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        e.generateName(nacos),
			Namespace:   nacos.Namespace,
			Labels:      labels,
			Annotations: nacos.Annotations,
		},
		Data: data,
	}
	myErrors.EnsureNormal(controllerutil.SetControllerReference(nacos, &cm, e.scheme))
	return &cm
}

func (e *KindClient) buildDefaultConfigMap(nacos *harmonycloudcnv1alpha1.Nacos) *v1.ConfigMap {
	labels := e.generateLabels(nacos.Name, NACOS)
	labels = e.MergeLabels(nacos.Labels, labels)
	data := make(map[string]string)

	// https://github.com/nacos-group/nacos-docker/blob/master/build/conf/application.properties
	data["application.properties"] = `# spring
	server.servlet.contextPath=${SERVER_SERVLET_CONTEXTPATH:/nacos}
	server.contextPath=/nacos
	server.port=${NACOS_APPLICATION_PORT:8848}
	spring.datasource.platform=${SPRING_DATASOURCE_PLATFORM:""}
	nacos.cmdb.dumpTaskInterval=3600
	nacos.cmdb.eventTaskInterval=10
	nacos.cmdb.labelTaskInterval=300
	nacos.cmdb.loadDataAtStart=false
	db.num=${MYSQL_DATABASE_NUM:1}
	db.url.0=jdbc:mysql://${MYSQL_SERVICE_HOST}:${MYSQL_SERVICE_PORT:3306}/${MYSQL_SERVICE_DB_NAME}?${MYSQL_SERVICE_DB_PARAM:characterEncoding=utf8&connectTimeout=1000&socketTimeout=3000&autoReconnect=true}
	db.url.1=jdbc:mysql://${MYSQL_SERVICE_HOST}:${MYSQL_SERVICE_PORT:3306}/${MYSQL_SERVICE_DB_NAME}?${MYSQL_SERVICE_DB_PARAM:characterEncoding=utf8&connectTimeout=1000&socketTimeout=3000&autoReconnect=true}
	db.user=${MYSQL_SERVICE_USER}
	db.password=${MYSQL_SERVICE_PASSWORD}
	### The auth system to use, currently only 'nacos' is supported:
	nacos.core.auth.system.type=${NACOS_AUTH_SYSTEM_TYPE:nacos}
	
	
	### The token expiration in seconds:
	nacos.core.auth.default.token.expire.seconds=${NACOS_AUTH_TOKEN_EXPIRE_SECONDS:18000}
	
	### The default token:
	nacos.core.auth.default.token.secret.key=${NACOS_AUTH_TOKEN:SecretKey012345678901234567890123456789012345678901234567890123456789}
	
	### Turn on/off caching of auth information. By turning on this switch, the update of auth information would have a 15 seconds delay.
	nacos.core.auth.caching.enabled=${NACOS_AUTH_CACHE_ENABLE:false}
	nacos.core.auth.enable.userAgentAuthWhite=${NACOS_AUTH_USER_AGENT_AUTH_WHITE_ENABLE:false}
	nacos.core.auth.server.identity.key=${NACOS_AUTH_IDENTITY_KEY:serverIdentity}
	nacos.core.auth.server.identity.value=${NACOS_AUTH_IDENTITY_VALUE:security}
	server.tomcat.accesslog.enabled=${TOMCAT_ACCESSLOG_ENABLED:false}
	server.tomcat.accesslog.pattern=%h %l %u %t "%r" %s %b %D
	# default current work dir
	server.tomcat.basedir=
	## spring security config
	### turn off security
	nacos.security.ignore.urls=${NACOS_SECURITY_IGNORE_URLS:/,/error,/**/*.css,/**/*.js,/**/*.html,/**/*.map,/**/*.svg,/**/*.png,/**/*.ico,/console-fe/public/**,/v1/auth/**,/v1/console/health/**,/actuator/**,/v1/console/server/**}
	# metrics for elastic search
	management.metrics.export.elastic.enabled=false
	management.metrics.export.influx.enabled=false
	
	nacos.naming.distro.taskDispatchThreadCount=10
	nacos.naming.distro.taskDispatchPeriod=200
	nacos.naming.distro.batchSyncKeyCount=1000
	nacos.naming.distro.initDataRatio=0.9
	nacos.naming.distro.syncRetryDelay=5000
	nacos.naming.data.warmup=true`

	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-default", e.generateName(nacos)),
			Namespace:   nacos.Namespace,
			Labels:      labels,
			Annotations: nacos.Annotations,
		},
		Data: data,
	}
	myErrors.EnsureNormal(controllerutil.SetControllerReference(nacos, &cm, e.scheme))
	return &cm
}

func (e *KindClient) buildStatefulsetCluster(nacos *harmonycloudcnv1alpha1.Nacos, ss *appv1.StatefulSet) *appv1.StatefulSet {
	ss.Spec.ServiceName = e.generateHeadlessSvcName(nacos)
	serivce := ""
	for i := 0; i < int(*nacos.Spec.Replicas); i++ {
		serivce = fmt.Sprintf("%v%v-%d.%v.%v.%v:%v ", serivce, e.generateName(nacos), i, e.generateHeadlessSvcName(nacos), nacos.Namespace, "svc.cluster.local", NACOS_PORT)
	}
	serivce = serivce[0 : len(serivce)-1]
	env := []v1.EnvVar{
		{
			Name:  "NACOS_SERVERS",
			Value: serivce,
		},
	}
	ss.Spec.Template.Spec.Containers[0].Env = append(ss.Spec.Template.Spec.Containers[0].Env, env...)
	return ss
}

func (e *KindClient) buildHeadlessServiceCluster(svc *v1.Service) *v1.Service {
	svc.Spec.ClusterIP = "None"
	return svc
}
