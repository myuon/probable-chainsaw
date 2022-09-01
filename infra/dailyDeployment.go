package infra

import (
	"github.com/myuon/probable-chainsaw/model"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type DailyDeployment struct {
	Date  string
	Count int
}

type DailyDeploymentCalculator struct {
	Db *gorm.DB
}

func (r DailyDeploymentCalculator) GetDailyDeployment(repositoryName string) ([]DailyDeployment, error) {
	deployments := []DailyDeployment{}

	if err := r.Db.
		Model(&model.DeployCommit{}).
		Where("repository_name = ?", repositoryName).
		Group("date(deployed_at, 'unixepoch')").
		Select("date(deployed_at, 'unixepoch') as date, count(hash) as count").
		Find(&deployments).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return deployments, nil
}
