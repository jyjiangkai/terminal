package eks

import (
	"github.com/gorilla/mux"
	k8s "k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"log"
	"net/http"
	ekstoken "terminal/pkg/eks/token"
	"terminal/pkg/kube"
	"terminal/pkg/session"
)

// Client contains all kube resource client
type Client struct {
	*PodBox
}

//type Client struct {
//	k8sClient kubernetes.Interface
//	config    *restclient.Config
//}

// NewClient new eke client
func NewClient(r *http.Request) (*Client, error) {
	pathParams := mux.Vars(r)
	clusterName := pathParams["clustername"]

	// Get project id form cookies, use ems api
	sessions, err := session.EmsSessionAuth(r)
	if err != nil {
		log.Printf("get session failed: %v\n", err)
		return nil, err
	}

	client, err := kube.NewClient()
	if err != nil {
		log.Printf("create eks client failed, error: %v", err)
		return nil, err
	}
	clusterList, err := client.ClusterBox.List(ClusterGVR, EksNamespace)
	if err != nil {
		log.Printf("get cluster info: cluster %s, projectID %s, error: %v", "redisnotdelete", sessions.ProjectID, err)
		return nil, err
	}
	// Get cluster info from cluster name and project id
	cluster, err := GetClusterInfo(clusterList, clusterName, sessions.ProjectID)
	if err != nil {
		log.Printf("get cluster info: cluster %s, projectID %s, error: %v", "redisnotdelete", sessions.ProjectID, err)
		return nil, err
	}

	// Get token
	token, err := ekstoken.GetToken(r, cluster, sessions.ProjectID)
	if err != nil {
		return nil, err
	}

	// New eks client with token
	vTrue := true
	kubeConfig := &k8s.ConfigFlags{
		APIServer:   &cluster.APIServerAddress,
		Insecure:    &vTrue,
		BearerToken: &token,
	}
	cfg, err := kubeConfig.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	cli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{&PodBox{clientset: cli, config: cfg}}, nil
}
