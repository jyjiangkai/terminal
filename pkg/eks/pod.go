package eks

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"terminal/pkg/terminal"
)

// PodBox provide functions for kubernetes pod.
type PodBox struct {
	clientset clientset.Interface
	config    *restclient.Config
}

// Get get specified pod in specified namespace.
func (c *PodBox) Get(name, namespace string) (*corev1.Pod, error) {
	opt := metav1.GetOptions{}
	return c.clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, opt)
}

// List list pods in specified namespace.
func (c *PodBox) List(namespace, labelSelector string) (*corev1.PodList, error) {
	opt := metav1.ListOptions{LabelSelector: labelSelector}
	return c.clientset.CoreV1().Pods(namespace).List(context.TODO(), opt)
}

// Exists check if pod exists.
func (c *PodBox) Exists(name, namespace string) (bool, error) {
	_, err := c.Get(name, namespace)
	if err == nil {
		return true, nil
	} else if apierrors.IsNotFound(err) {
		return false, nil
	}
	return false, err
}

// Create creates a pod
func (c *PodBox) Create(pod *corev1.Pod, namespace string) (*corev1.Pod, error) {
	opt := metav1.CreateOptions{}
	return c.clientset.CoreV1().Pods(namespace).Create(context.TODO(), pod, opt)
}

// Watch watch pod in specified namespace with timeoutSeconds
func (c *PodBox) Watch(namespace string, timeoutSeconds *int64, labelSelector string) (watch.Interface, error) {
	opt := metav1.ListOptions{TimeoutSeconds: timeoutSeconds, LabelSelector: labelSelector}
	return c.clientset.CoreV1().Pods(namespace).Watch(context.TODO(), opt)
}

// WatchPod watch specified pod in specified namespace with timeoutSeconds
func (c *PodBox) WatchPod(namespace, podName string, timeoutSeconds *int64) (watch.Interface, error) {
	pod, err := c.Get(podName, namespace)
	if err != nil {
		return nil, err
	}
	opt := metav1.ListOptions{
		TimeoutSeconds:  timeoutSeconds,
		FieldSelector:   fmt.Sprintf("metadata.name=%s", podName),
		ResourceVersion: pod.ResourceVersion,
	}
	w, err := c.clientset.CoreV1().Pods(namespace).Watch(context.TODO(), opt)
	return w, err
}

// Exec exec into a pod
func (c *PodBox) Exec(cmd []string, ptyHandler terminal.PtyHandler, namespace, podName, containerName string) error {
	defer func() {
		ptyHandler.Done()
	}()

	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   cmd,
		Stdin:     !(ptyHandler.Stdin() == nil),
		Stdout:    !(ptyHandler.Stdout() == nil),
		Stderr:    !(ptyHandler.Stderr() == nil),
		TTY:       ptyHandler.Tty(),
	}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return err
	}
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:             ptyHandler.Stdin(),
		Stdout:            ptyHandler.Stdout(),
		Stderr:            ptyHandler.Stderr(),
		TerminalSizeQueue: ptyHandler,
		Tty:               ptyHandler.Tty(),
	})
	return err
}

// Delete delete pod
func (c *PodBox) Delete(name, namespace string) error {
	opt := metav1.DeleteOptions{}
	return c.clientset.CoreV1().Pods(namespace).Delete(context.TODO(), name, opt)
}
