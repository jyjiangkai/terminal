package eks

import (
	"context"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

var (
	clusterGVR = schema.GroupVersionResource{
		Group:    "ecns.easystack.com",
		Version:  "v1",
		Resource: "clusters",
	}
)

const (
	EKS_NAMESPACE = "eks"
	EKS_TYPE      = "EKS"
)

type EOSClient struct {
	client dynamic.Interface
}

func NewEOSclient(kubeconfig string) (*EOSClient, error) {
	var config *rest.Config
	var err error

	if kubeconfig == "" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if err != nil {
		return nil, err
	}

	client, err := dynamic.NewForConfig(config)

	if err != nil {
		return nil, err
	}

	return &EOSClient{client: client}, nil
}

// EKSCluster struct is generated from json
type EKSCluster struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		CreationTimestamp time.Time `json:"creationTimestamp"`
		Generation        int       `json:"generation"`
		Labels            struct {
			Clustername string `json:"clustername"`
		} `json:"labels"`
		Name            string `json:"name"`
		Namespace       string `json:"namespace"`
		ResourceVersion string `json:"resourceVersion"`
		SelfLink        string `json:"selfLink"`
		UID             string `json:"uid"`
	} `json:"metadata"`
	Spec struct {
		Architecture string `json:"architecture"`
		Clusterid    string `json:"clusterid"`
		Eks          struct {
			APIAddress   string `json:"api_address"`
			EksClusterid string `json:"eks_clusterid"`
			EksName      string `json:"eks_name"`
			EksStackid   string `json:"eks_stackid"`
			EksStatus    string `json:"eks_status"`
		} `json:"eks"`
		Host       string   `json:"host"`
		NodesCount int      `json:"nodes_count"`
		Projects   []string `json:"projects"`
		Type       string   `json:"type"`
		Version    string   `json:"version"`
	} `json:"spec"`
	Status struct {
		ClusterStatus     string `json:"cluster_status"`
		HasReconciledOnce bool   `json:"has_reconciled_once"`
		Nodes             []struct {
			ComponentStatus struct {
			} `json:"component_status"`
			NodeName   string `json:"node_name"`
			NodeRole   string `json:"node_role"`
			NodeStatus string `json:"node_status"`
		} `json:"nodes"`
	} `json:"status"`
}

func (cli *EOSClient) Clusters(projectID string) ([]EKSCluster, error) {
	list, err := cli.client.Resource(clusterGVR).Namespace(EKS_NAMESPACE).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []EKSCluster

	for _, c := range list.Items {
		jsonBytes, err := c.MarshalJSON()
		if err != nil {
			return nil, err
		}
		eksCluster := EKSCluster{}
		err = json.Unmarshal(jsonBytes, &eksCluster)
		if err != nil {
			return nil, err
		}
		if eksCluster.Spec.Type != EKS_TYPE {
			continue
		}
		if projectID == "" {
			result = append(result, eksCluster)
		} else {
			for _, p := range eksCluster.Spec.Projects {
				if p == projectID {
					result = append(result, eksCluster)
				}
			}
		}
	}

	return result, nil
}
