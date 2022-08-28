package infra

import (
	"github.com/myuon/probable-chainsaw/model"
	"gorm.io/gorm"
)

type DailyDeployment struct {
	Date  string
	Count int
}

type DailyDeploymentCalculator struct {
	Db *gorm.DB
}

func (r DailyDeploymentCalculator) GetDailyDeployment() ([]DailyDeployment, error) {
	deployments := []DailyDeployment{}

	if err := r.Db.
		Model(&model.Deployment{}).
		Group("date(deployed_time)").
		Select("date(deployed_time) as date, count(id) as count").
		Find(&deployments).Error; err != nil {
		return nil, err
	}

	return deployments, nil
}
