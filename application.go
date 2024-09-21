package tuikit

type Application struct {
	Name string

	stateKey   string
	stateVal   string
	notice     string
	loadingMsg string
}

type ApplicationOption func(*Application)

func NewApplication(name string, opts ...ApplicationOption) *Application {
	if name == "" {
		name = "tuikit"
	}
	a := &Application{Name: name}
	for _, opt := range opts {
		opt(a)
	}
	if a.loadingMsg == "" {
		a.loadingMsg = "loading..."
	}
	return a
}

func WithStateKey(key string) ApplicationOption {
	return func(a *Application) {
		a.stateKey = key
	}
}

func WithStateVal(val string) ApplicationOption {
	return func(a *Application) {
		a.stateVal = val
	}
}

func WithState(key, val string) ApplicationOption {
	return func(a *Application) {
		a.stateKey = key
		a.stateVal = val
	}
}

func WithNotice(notice string) ApplicationOption {
	return func(a *Application) {
		a.notice = notice
	}
}

func WithLoadingMsg(msg string) ApplicationOption {
	return func(a *Application) {
		a.loadingMsg = msg
	}
}
