package eks

import (
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"time"
)

var (
	ClusterGVR = schema.GroupVersionResource{
		Group:    "ecns.easystack.com",
		Version:  "v1",
		Resource: "clusters",
	}
)

const (
	EksNamespace = "eks"
	EksType      = "EKS"
)

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

// Cluster cluster
//
// swagger:model Cluster
type Cluster struct {

	// api server address
	// Required: true
	APIServerAddress string `json:"apiServerAddress"`

	// healthy
	// Required: true
	Healthy bool `json:"healthy"`

	// name
	// Required: true
	Name string `json:"name"`

	// owned by current user
	// Required: true
	OwnedByCurrentUser bool `json:"owned_by_current_user"`

	// project ID
	// Required: true
	ProjectID string `json:"projectID"`

	// status
	// Required: true
	Status string `json:"status"`
}

func GetClusterInfo(clusterList *unstructured.UnstructuredList, clusterName, projectID string) (*Cluster, error) {
	for _, cluster := range clusterList.Items {
		jsonBytes, err := cluster.MarshalJSON()
		if err != nil {
			return nil, err
		}
		eksCluster := EKSCluster{}
		err = json.Unmarshal(jsonBytes, &eksCluster)
		if err != nil {
			return nil, err
		}
		if eksCluster.Spec.Type != EksType {
			continue
		}

		name := eksCluster.Spec.Eks.EksName
		projects := eksCluster.Spec.Projects

		if len(projects) > 0 && projects[0] == projectID && name == clusterName {
			retc := &Cluster{
				APIServerAddress:   eksCluster.Spec.Eks.APIAddress,
				Name:               eksCluster.Spec.Eks.EksName,
				ProjectID:          eksCluster.Spec.Projects[0],
				OwnedByCurrentUser: true,
				Status:             eksCluster.Status.ClusterStatus,
			}
			switch eksCluster.Status.ClusterStatus {
			case "Healthy", "UPDATE_IN_PROGRESS", "UPDATE_FAILED", "Warning":
				retc.Healthy = true
			default:
				retc.Healthy = false
			}
			return retc, nil
		}
	}
	return nil, fmt.Errorf("cluster %s not found", clusterName)
}
