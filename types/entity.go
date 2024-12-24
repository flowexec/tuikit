package types

import "fmt"

const (
	EntityFormatJSON     Format = "json"
	EntityFormatYAML     Format = "yaml"
	EntityFormatDocument Format = "md"
)

type Entity interface {
	YAML() (string, error)
	JSON() (string, error)
	Markdown() string
}

type EntityInfo struct {
	Header    string
	SubHeader string
	Desc      string
	ID        string
}

func (i *EntityInfo) Title() string {
	title := i.Header
	if i.SubHeader != "" {
		title += fmt.Sprintf(" (%s)", i.SubHeader)
	}
	return title
}

func (i *EntityInfo) Description() string { return i.Desc }
func (i *EntityInfo) FilterValue() string { return i.ID }
