package types

type Entity interface {
	YAML() (string, error)
	JSON(formatted bool) (string, error)
	Markdown() string
}
