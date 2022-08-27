package model

import "github.com/oklog/ulid/v2"

type DeploymentId string

func NewDeploymentId() DeploymentId {
	return DeploymentId(ulid.Make().String())
}

type Deployment struct {
	Id           DeploymentId `gorm:"primaryKey"`
	DeployedTime string
	CommitHash   string
}
