package types

type Entity interface {
	YAML() (string, error)
	JSON() (string, error)
	Markdown() string
}
