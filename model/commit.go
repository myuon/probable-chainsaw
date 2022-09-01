package model

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
