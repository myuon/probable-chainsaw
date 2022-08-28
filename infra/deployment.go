package infra

import (
	"github.com/myuon/probable-chainsaw/model"
	"gorm.io/gorm"
	"time"
)

type DeploymentRepository struct {
	Db *gorm.DB
}

func (r DeploymentRepository) ResetTable() error {
	if err := r.Db.Migrator().DropTable(&model.Deployment{}); err != nil {
		return err
	}
	if err := r.Db.AutoMigrate(&model.Deployment{}); err != nil {
		return err
	}

	return nil
}

func (r DeploymentRepository) FindByDeployedAt(start time.Time, end time.Time) ([]model.Deployment, error) {
	rs := []model.Deployment{}
	if err := r.Db.Where("deployed_time >= ? AND deployed_time < ?", start.String(), end.String()).Find(&rs).Error; err != nil {
		return nil, err
	}

	return rs, nil
}

func (r DeploymentRepository) Create(t model.Deployment) error {
	if err := r.Db.Create(&t).Error; err != nil {
		return err
	}

	return nil
}
