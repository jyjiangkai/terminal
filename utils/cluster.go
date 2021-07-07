package utils

import (
	"terminal/models"
	"terminal/pkg/eks"
	"fmt"
	"sort"
)

type ClustersController struct{}

type eksClusters []*models.Cluster

func (c eksClusters) Len() int { return len(c) }

func (c eksClusters) Less(i, j int) bool {
	if *c[i].Healthy == true && *c[j].Healthy == false {
		return true
	}
	return false
}

func (c eksClusters) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// 获取 eos 集群上所有 cluster crd的实例列表，用projectID过滤。如果 是 getAll，则返回所有集群
func allClusters(projectID string, getAll bool) ([]*models.Cluster, error) {
	eosClient, err := eks.NewEOSclient("")
	if err != nil {
		return nil, err
	}

	p := projectID
	if getAll {
		p = ""
	}
	clusters, err := eosClient.Clusters(p)
	if err != nil {
		return nil, err
	}

	cs := make([]*models.Cluster, len(clusters))

	for i, v := range clusters {
		pid := ""
		if len(v.Spec.Projects) > 0 {
			pid = v.Spec.Projects[0]
		}
		cs[i] = &models.Cluster{
			APIServerAddress:   pString(v.Spec.Eks.APIAddress),
			Name:               pString(v.Spec.Eks.EksName),
			ProjectID:          &pid,
			OwnedByCurrentUser: pBool(pid == projectID),
			Status:             pString(v.Status.ClusterStatus),
		}
		switch v.Status.ClusterStatus {
		case "Healthy", "UPDATE_IN_PROGRESS", "UPDATE_FAILED", "Warning":
			cs[i].Healthy = pBool(true)
		default:
			cs[i].Healthy = pBool(false)
		}
	}
	sort.Sort(eksClusters(cs))
	return cs, nil
}

func GetClusterInfo(name, projectID string) (*models.Cluster, error) {
	clusters, err := allClusters(projectID, true)
	if err != nil {
		return nil, err
	}

	var cluster *models.Cluster
	var count int

	for _, c := range clusters {
		if *c.Name == name {
			cluster = c
			count++
		}
	}
	if count == 1 {
		return cluster, nil
	}

	// 集群可能出现重名的情况，所以如果找到一个以上重名的集群，就要再比对一次项目ID
	for _, c := range clusters {
		if *c.Name == name && *c.ProjectID == projectID {
			return c, nil
		}
	}

	return nil, fmt.Errorf("cluster %s not found", name)
}
