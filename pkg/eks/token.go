package eks

import (
	k8s "k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"terminal/pkg/eks/cache"
)

// validateToken
// use the token to construct eks client and access any k8s interface.
// if the access is successful, the token is valid.
func validateToken(cluster *Cluster, token string) error {
	// construct eks client
	vTrue := true
	kubeConfig := &k8s.ConfigFlags{
		APIServer:   &cluster.APIServerAddress,
		Insecure:    &vTrue,
		BearerToken: &token,
	}

	rest, err := kubeConfig.ToRESTConfig()
	if err != nil {
		return err
	}

	client, err := kubernetes.NewForConfig(rest)
	if err != nil {
		return err
	}

	_, err = client.Discovery().ServerVersion()
	return err
}

// GetToken
func GetToken(r *http.Request, cluster *Cluster, projectID string) (string, error) {
	// step1: get token from cache
	token, err := cache.GetTokenFromCache(r.Cookies(), projectID)
	if err != nil {
		return "", err
	}
	// step2: vaildate token, if valid, return it
	err = validateToken(cluster, token)
	if err == nil {
		return token, nil
	}

	// step3: if token is invaild, query token from openstack again
	token, err = cache.GetTokenFromOpenstack(r.Cookies(), projectID)
	if err != nil {
		return "", err
	}

	// step4: vaildate token again
	err = validateToken(cluster, token)
	if err != nil {
		return "", err
	}
	return token, err
}
