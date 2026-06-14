// Negative (edge): the composition root — the application's New() per the template.
package api

type Application struct{}

func New() *Application {
	return &Application{}
}
