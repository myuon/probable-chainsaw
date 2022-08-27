package model

type Commit struct {
	Hash       string `sql:"primaryKey"`
	AuthorName string
	CreatedAt  int64
	DeployTag  string
}
