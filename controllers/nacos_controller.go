/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"reflect"

	corev1 "k8s.io/api/core/v1"

	"harmonycloud.cn/nacos-operator/pkg/service/operator"

	"github.com/go-logr/logr"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	harmonycloudcnv1alpha1 "harmonycloud.cn/nacos-operator/api/v1alpha1"

	myErrors "harmonycloud.cn/nacos-operator/pkg/errors"
)

// NacosReconciler reconciles a Nacos object
type NacosReconciler struct {
	client.Client
	Log            logr.Logger
	Scheme         *runtime.Scheme
	OperaterClient *operator.OperatorClient
}

// +kubebuilder:rbac:groups=harmonycloud.cn.harmonycloud.cn,resources=nacos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=harmonycloud.cn.harmonycloud.cn,resources=nacos/status,verbs=get;update;patch
type reconcileFun func(nacos *harmonycloudcnv1alpha1.Nacos)

func (r *NacosReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("nacos", req.NamespacedName)

	instance := &harmonycloudcnv1alpha1.Nacos{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// 处理全局异常处理中的异常
	defer func() {
		if err := recover(); err != nil {
			r.Log.Error(err.(error), "error")
		}
	}()

	// 全局处理异常
	defer func() {
		if err := recover(); err != nil {
			r.globalExceptHandle(err)
		}
	}()

	for _, fun := range []reconcileFun{
		// 保证资源能够创建
		r.OperaterClient.MakeEnsure,
		// 检查并修复
		r.OperaterClient.CheckAndMakeHeal,
		// 保存状态
		r.OperaterClient.UpdateStatus,
	} {
		fun(instance)
	}

	return ctrl.Result{}, nil
}

func (r *NacosReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&harmonycloudcnv1alpha1.Nacos{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}

// 全局异常处理
func (r *NacosReconciler) globalExceptHandle(err interface{}) {
	if reflect.TypeOf(err) == reflect.TypeOf(myErrors.NewErrMsg("")) {
		myerr := err.(*myErrors.Err)
		klog.Warningf("painc msg[%s] code[%d]", myerr.Msg, myerr.Code)
	} else {
		r.Log.Error(err.(error), "error")
	}
}
