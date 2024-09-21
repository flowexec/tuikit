package types

type NoticeLevel string
type Format string

type KeyCallback struct {
	Key      string
	Label    string
	Callback func() error
}
