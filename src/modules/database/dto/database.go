package dto

type ListDatabasesInput struct{}

type Database struct {
	Name    string `json:"name"`
	Default bool   `json:"default"`
}

type ListDatabasesOutput struct {
	Default   string     `json:"default"`
	Databases []Database `json:"databases"`
}
