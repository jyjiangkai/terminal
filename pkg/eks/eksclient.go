package eks

import (
	"k8s.io/apimachinery/pkg/version"
	k8s "k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

type EKSClient struct {
	k8sClient kubernetes.Interface
	config    *restclient.Config
}

func NewEKSClient(apiserver, token string) (*EKSClient, error) {
	vTrue := true
	kubeConfig := &k8s.ConfigFlags{
		APIServer:   &apiserver,
		Insecure:    &vTrue,
		BearerToken: &token,
	}
	rest, err := kubeConfig.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	cliSet, err := kubernetes.NewForConfig(rest)
	if err != nil {
		return nil, err
	}

	return &EKSClient{k8sClient: cliSet, config: rest}, nil
}

func (c *EKSClient) ServerVersion() (*version.Info, error) {
	return c.k8sClient.Discovery().ServerVersion()
}
