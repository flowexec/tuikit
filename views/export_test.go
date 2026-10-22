package views

// Stub the native clipboard for the whole test binary so tests never touch the real one.
func init() {
	writeNativeClipboard = func(string) error { return nil }
}

// SetNativeClipboardForTest replaces the native clipboard writer and returns a func that restores it.
func SetNativeClipboardForTest(f func(string) error) (restore func()) {
	orig := writeNativeClipboard
	writeNativeClipboard = f
	return func() { writeNativeClipboard = orig }
}
