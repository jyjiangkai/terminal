// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

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
