package operator

import harmonycloudcnv1alpha1 "harmonycloud.cn/nacos-operator/api/v1alpha1"

type ICheckClient interface {
}

type CheckClient struct {
}

func (c CheckClient) CheckNum(nacos *harmonycloudcnv1alpha1.Nacos) {

}
