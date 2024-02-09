package components

import (
	"fmt"

	"github.com/jahvon/tuikit/styles"
)

type Header struct {
	Styles      styles.Theme
	Name        string
	CtxKey      string
	CtxVal      string
	Notice      string
	NoticeLevel NoticeLevel
}

func (h *Header) View() string {
	header := h.Styles.RenderBrand(h.Name)
	header += ContextStr(h.CtxKey, h.CtxVal, h.Styles)
	header += NoticeStr(h.Notice, h.NoticeLevel, h.Styles)
	return header
}

func (h *Header) Print() {
	fmt.Println(h.View() + "\n")
}

func ContextStr(label, val string, styles styles.Theme) string {
	if label == "" {
		label = "ctx"
	}
	if val == "" {
		val = styles.RenderUnknown("unk")
	}
	return styles.RenderContext(label, val)
}

func NoticeStr(notice string, lvl NoticeLevel, styles styles.Theme) string {
	if notice == "" {
		return ""
	}

	switch lvl {
	case NoticeLevelInfo:
		return styles.RenderInContainer(styles.RenderSuccess(notice))
	case NoticeLevelWarning:
		return styles.RenderInContainer(styles.RenderWarning(notice))
	case NoticeLevelError:
		return styles.RenderInContainer(styles.RenderError(notice))
	default:
		return styles.RenderInContainer(styles.RenderUnknown(notice))
	}
}
