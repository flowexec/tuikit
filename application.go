package tuikit

type Application struct {
	Name    string
	Version string

	stateKey   string
	stateVal   string
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

func WithVersion(version string) ApplicationOption {
	return func(a *Application) {
		a.Version = version
	}
}

func WithLoadingMsg(msg string) ApplicationOption {
	return func(a *Application) {
		a.loadingMsg = msg
	}
}
