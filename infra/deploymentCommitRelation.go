package infra

import (
	"github.com/myuon/probable-chainsaw/model"
	"gorm.io/gorm"
)

type DeploymentCommitRelation struct {
	DeploymentId model.DeploymentId `gorm:"primaryKey"`
	CommitHash   string             `gorm:"primaryKey"`
	LeadTime     int64
}

type DeploymentCommitRelationRepository struct {
	Db *gorm.DB
}

func (r DeploymentCommitRelationRepository) ResetTable() error {
	if err := r.Db.Migrator().DropTable(&DeploymentCommitRelation{}); err != nil {
		return err
	}
	if err := r.Db.AutoMigrate(&DeploymentCommitRelation{}); err != nil {
		return err
	}

	return nil
}

func (r DeploymentCommitRelationRepository) FindByDeploymentId(deploymentId model.DeploymentId) ([]DeploymentCommitRelation, error) {
	rs := []DeploymentCommitRelation{}
	if err := r.Db.Where("deployment_id = ?", deploymentId).Find(&rs).Error; err != nil {
		return nil, err
	}

	return rs, nil
}

func (r DeploymentCommitRelationRepository) Create(relations []DeploymentCommitRelation) error {
	if err := r.Db.Create(&relations).Error; err != nil {
		return err
	}

	return nil
}

func (r DeploymentCommitRelationRepository) Save(relations []DeploymentCommitRelation) error {
	if err := r.Db.Save(&relations).Error; err != nil {
		return err
	}

	return nil
}
