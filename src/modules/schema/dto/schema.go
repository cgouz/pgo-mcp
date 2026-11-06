package dto

type ListSchemasInput struct{}

type ListSchemasOutput struct {
	Schemas []string `json:"schemas"`
}
