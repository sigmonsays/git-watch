package watch

type Watcher interface {
	Start() error
	Stop() error
	Stopped() bool
}
