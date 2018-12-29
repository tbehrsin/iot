package errors

import (
	"encoding/json"
	"net/http"
)

func APIJSON(w http.ResponseWriter, body interface{}) {
	wrapper := map[string]interface{}{}
	wrapper["body"] = body

	if data, err := json.Marshal(wrapper); err != nil {
		NewInternalServerError(err).Write(w)
	} else {
		w.Header()["Content-Type"] = []string{"application/json"}
		w.Header()["Cache-Control"] = []string{"no-cache"}
		// w.Header()["Expires"] = []string{"Thu, 01 Jan 1970 00:00:00 GMT"}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}
