package types

import (
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"

	"github.com/jahvon/tuikit/types"
)

type Echo struct {
	Content string
}

func (m *Echo) Init() tea.Cmd {
	return nil
}

func (m *Echo) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *Echo) View() string {
	if m.Content == "" {
		return "Hello, World!"
	}
	return m.Content
}

type Thing struct {
	Name string `json:"data"`
	Type string `json:"type"`
}

func (t *Thing) YAML() (string, error) {
	v, err := yaml.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func (t *Thing) JSON() (string, error) {
	v, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func (t *Thing) Markdown() string {
	return fmt.Sprintf("# %s\n\nI am a(n) **%s.**", t.Name, t.Type)
}

type ThingList struct {
	thingType string
	items     []*types.EntityInfo
}

func NewThingList(thingType string, things ...*types.EntityInfo) *ThingList {
	return &ThingList{
		thingType: thingType,
		items:     things,
	}
}

func (t *ThingList) Items() []*types.EntityInfo {
	return t.items
}

func (t *ThingList) YAML() (string, error) {
	v, err := yaml.Marshal(t.items)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func (t *ThingList) JSON() (string, error) {
	v, err := json.Marshal(t.items)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func (t *ThingList) Singular() string {
	if t.thingType != "" {
		return t.thingType
	}
	return "thing"
}

func (t *ThingList) Plural() string {
	if t.thingType != "" {
		return t.thingType + "s"
	}
	return "things"
}
