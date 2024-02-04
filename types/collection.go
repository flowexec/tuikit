package types

import "fmt"

type CollectionItem struct {
	Header    string
	SubHeader string
	Desc      string
}

func (i *CollectionItem) Title() string {
	title := i.Header
	if i.SubHeader != "" {
		title += fmt.Sprintf(" (%s)", i.SubHeader)
	}
	return title
}

func (i *CollectionItem) Description() string { return i.Desc }
func (i *CollectionItem) FilterValue() string { return i.Header }

type Collection interface {
	Items() []*CollectionItem
	YAML() (string, error)
	JSON() (string, error)
	Singular() string
	Plural() string
}
