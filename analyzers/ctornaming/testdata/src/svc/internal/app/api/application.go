// Негатив (граница): composition root — New() приложения по шаблону.
package api

type Application struct{}

func New() *Application {
	return &Application{}
}
