package infra

import (
	"github.com/myuon/probable-chainsaw/model"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type DeployCommitRepository struct {
	Db *gorm.DB
}

func (r DeployCommitRepository) ResetTable() error {
	if err := r.Db.Migrator().DropTable(&model.DeployCommit{}); err != nil {
		return errors.WithStack(err)
	}

	if err := r.Db.AutoMigrate(&model.DeployCommit{}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r DeployCommitRepository) FindBetweenDeployedAt(start int64, end int64) ([]model.DeployCommit, error) {
	rs := []model.DeployCommit{}
	if err := r.Db.Where("deployed_at >= ? AND deployed_at < ?", start, end).Find(&rs).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return rs, nil
}

func (r DeployCommitRepository) Create(commits []model.DeployCommit) error {
	if err := r.Db.Create(&commits).Error; err != nil {
		return errors.WithStack(err)
	}

	return nil
}
