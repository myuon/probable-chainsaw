package infra

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type DeployCommitRelation struct {
	DeployHash string
	CommitHash string `gorm:"primaryKey"`
}

type DeployCommitRelationRepository struct {
	Db *gorm.DB
}

func (r DeployCommitRelationRepository) ResetTable() error {
	if err := r.Db.Migrator().DropTable(&DeployCommitRelation{}); err != nil {
		return errors.WithStack(err)
	}
	if err := r.Db.AutoMigrate(&DeployCommitRelation{}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r DeployCommitRelationRepository) FindByDeployHash(deployHash string) ([]DeployCommitRelation, error) {
	rs := []DeployCommitRelation{}
	if err := r.Db.Where("deploy_hash = ?", deployHash).Find(&rs).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return rs, nil
}

func (r DeployCommitRelationRepository) Create(relations []DeployCommitRelation) error {
	if err := r.Db.Create(&relations).Error; err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r DeployCommitRelationRepository) Save(relations []DeployCommitRelation) error {
	if err := r.Db.Save(&relations).Error; err != nil {
		return errors.WithStack(err)
	}

	return nil
}

type DeployLeadTime struct {
	DeployHash string
	LeadTime   int
}

func (r DeployCommitRelationRepository) GetLeadTimeForDeploys(deployCommitHashes []string) ([]DeployLeadTime, error) {
	ms := []DeployLeadTime{}

	if err := r.Db.
		Model(&DeployCommitRelation{}).
		Joins("inner join commits on commits.hash = deploy_commit_relations.commit_hash").
		Joins("inner join deploy_commits on deploy_commits.hash = deploy_commit_relations.deploy_hash").
		Where("deploy_hash in ?", deployCommitHashes).
		Group("deploy_hash").
		Select("deploy_hash, max(deployed_at - created_at) as lead_time").
		Find(&ms).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return ms, nil
}
