package lib

import "fmt"

type Widget struct {
	Name string
}

type Runner interface {
	Run() error
}

func NewWidget(name string) *Widget {
	return &Widget{Name: name}
}

func (w *Widget) Run() error {
	if w.Name == "" {
		return fmt.Errorf("missing name")
	}
	return nil
}

func (w *Widget) String() string {
	return fmt.Sprintf("Widget(%s)", w.Name)
}

func UseWidget(r Runner) error {
	return r.Run()
}

func WrapWidget(w *Widget) Runner {
	return widgetRunner{Widget: w}
}

type widgetRunner struct {
	*Widget
}

func (wr widgetRunner) Run() error {
	return wr.Widget.Run()
}
