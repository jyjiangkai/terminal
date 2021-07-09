package kube

import (
	kubeclient "terminal/pkg/client"
	"terminal/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

var deletePolicy = metav1.DeletePropagationForeground
var commonDeleteOpt = metav1.DeleteOptions{
	GracePeriodSeconds: utils.Int64Ptr(0),
	PropagationPolicy:  &deletePolicy,
}

// Client contains all kube resource client
type Client struct {
	*PodBox
	*ClusterBox
}

// Client contains all kube resource client
//type Client struct {
//	clientk8s kubernetes.Interface
//	config    restclient.Config
//	clientdyc dynamic.Interface
//}

// NewClient get all kube resource client.
func NewClient() (*Client, error) {
	kubeClient := kubeclient.KubeClientset()
	cfg, err := kubeclient.Config()
	dycClient := kubeclient.DynamicClientset()
	if err != nil {
		return nil, err
	}
	client := &Client{
		&PodBox{clientset: kubeClient, config: cfg},
		&ClusterBox{clientset: dycClient},
	}
	return client, nil
}

// DecodeKubeObj decode kubernetes object from yaml
func DecodeKubeObj(yml []byte) (k8sruntime.Object, *schema.GroupVersionKind, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	return decode(yml, nil, nil)
}
