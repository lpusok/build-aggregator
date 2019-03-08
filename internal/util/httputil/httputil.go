package httputil

import (
	"fmt"
	"net/http"
)

// Check returns false and an error, if the response is not
// of the expected http response type or if the error passed
// is not nil (e.g. because of an error sending the request)
func Check(res *http.Response, reqErr error, statusCode int) (bool, error) {
	if res == nil || reqErr != nil {
		return false, reqErr
	}

	if res.StatusCode != statusCode {
		return false, fmt.Errorf("http response %d %s", res.StatusCode, res.Status)
	}

	return true, nil
}
