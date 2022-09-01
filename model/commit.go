package model

import "time"

type Commit struct {
	Hash           string `gorm:"primaryKey"`
	AuthorName     string
	CreatedAt      int64
	DeployTag      string
	Parent         string
	RepositoryName string
}

type DeployCommit struct {
	Hash           string `gorm:"primaryKey"`
	AuthorName     string
	DeployedAt     int64
	PreviousHash   string
	RepositoryName string
	LeadTime       int
}

type DeployCommits []DeployCommit

func (ds DeployCommits) Hashes() []string {
	hashes := []string{}
	for _, d := range ds {
		hashes = append(hashes, d.Hash)
	}
	return hashes
}

func (ds DeployCommits) LeadTimeAvg() int {
	if len(ds) == 0 {
		return 0
	}

	sum := 0
	for _, d := range ds {
		sum += d.LeadTime
	}
	return sum / len(ds)
}

func (ds DeployCommits) DeployDailyMap() map[string]int {
	deployMap := map[string]int{}
	for _, d := range ds {
		c := time.Unix(d.DeployedAt, 0).Format("2006-01-02")
		if _, ok := deployMap[c]; !ok {
			deployMap[c] = 0
		}

		deployMap[c] += 1
	}

	return deployMap
}
