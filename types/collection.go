package types

const (
	CollectionFormatList Format = "list"
	CollectionFormatJSON Format = "json"
	CollectionFormatYAML Format = "yaml"
)

type Collection interface {
	Items() []*EntityInfo
	YAML() (string, error)
	JSON() (string, error)
	Singular() string
	Plural() string
}
