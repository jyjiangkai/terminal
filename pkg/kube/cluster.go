package kube

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// ClusterBox provide functions for kubernetes pod.
type ClusterBox struct {
	clientset dynamic.Interface
}

func (c *ClusterBox) List(resource schema.GroupVersionResource, namespace string) (*unstructured.UnstructuredList, error) {
	ctx := context.Background()
	opt := metav1.ListOptions{}
	return c.clientset.Resource(resource).Namespace(namespace).List(ctx, opt)
}
