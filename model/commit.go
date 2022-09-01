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
}
