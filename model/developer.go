package model

type Developer struct {
	GithubHandle string   `json:"github_id" bson:"github_id"`
	Blog         string   `json:"blog"`
	Skills       []string `json:"skills"`
}
