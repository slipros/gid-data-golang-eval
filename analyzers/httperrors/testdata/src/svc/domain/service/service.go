// Inapplicable: outside server/http the rule does not apply.
package service

import "net/http"

func Helper(w http.ResponseWriter, err error) {}
