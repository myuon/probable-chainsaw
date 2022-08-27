package model

type Commit struct {
	Hash       string `gorm:"primaryKey"`
	AuthorName string
	CreatedAt  int64
	DeployTag  string
	Parent     string
}
