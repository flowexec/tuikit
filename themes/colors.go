package themes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ColorPalette struct {
	Primary   string `json:"primary" yaml:"primary"`
	Secondary string `json:"secondary" yaml:"secondary"`
	Tertiary  string `json:"tertiary" yaml:"tertiary"`
	Success   string `json:"success" yaml:"success"`
	Warning   string `json:"warning" yaml:"warning"`
	Error     string `json:"error" yaml:"error"`
	Info      string `json:"info" yaml:"info"`
	Body      string `json:"body" yaml:"body"`
	Emphasis  string `json:"emphasis" yaml:"emphasis"`
	Border    string `json:"border" yaml:"border"`

	Black string `json:"black" yaml:"black"`
	White string `json:"white" yaml:"white"`
	Gray  string `json:"gray" yaml:"gray"`

	// see https://github.com/alecthomas/chroma
	ChromaCodeStyle string `json:"chromaCodeStyle" yaml:"chromaCodeStyle"`
}

func ReadColorPalette(file string) (*ColorPalette, error) {
	cp := &ColorPalette{}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var data []byte
	if _, err = f.Read(data); err != nil {
		return nil, err
	}

	ext := filepath.Ext(file)
	switch strings.ToLower(ext) {
	case ".json":
		if err = json.Unmarshal(data, cp); err != nil {
			return nil, err
		}
	case ".yaml", ".yml":
		if err = yaml.Unmarshal(data, cp); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	return cp, nil
}

func WithDefaultColors(orig ColorPalette) ColorPalette {
	cp := orig
	if cp.Primary == "" {
		cp.Primary = "#7FBBB3"
	}
	if cp.Secondary == "" {
		cp.Secondary = "#83C092"
	}
	if cp.Tertiary == "" {
		cp.Tertiary = "#D699B6"
	}
	if cp.Success == "" {
		cp.Success = "#8DA101"
	}
	if cp.Warning == "" {
		cp.Warning = "#5C6A72"
	}
	if cp.Error == "" {
		cp.Error = "#F85552"
	}
	if cp.Info == "" {
		cp.Info = "#3A94C5"
	}
	if cp.Body == "" {
		cp.Body = "#D3C6AA"
	}
	if cp.Emphasis == "" {
		cp.Emphasis = "#E67E80"
	}
	if cp.Border == "" {
		cp.Border = "#5C6A72"
	}
	if cp.Black == "" {
		cp.Black = "#343F44"
	}
	if cp.White == "" {
		cp.White = "#DFDDC8"
	}
	if cp.Gray == "" {
		cp.Gray = "#5C6A72"
	}
	if cp.ChromaCodeStyle == "" {
		cp.ChromaCodeStyle = "friendly"
	}
}
