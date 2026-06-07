// Неприменимость: вне server/http правило не действует.
package service

import "net/http"

func Helper(w http.ResponseWriter, err error) {}
