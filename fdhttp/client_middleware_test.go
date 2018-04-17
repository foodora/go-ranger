package fdhttp_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/foodora/go-ranger/fdhttp"
	"github.com/stretchr/testify/assert"
)

func newClientMiddleware(called *bool) fdhttp.ClientMiddleware {
	return func(next fdhttp.Doer) fdhttp.Doer {
		return fdhttp.DoerFunc(func(req *http.Request) (*http.Response, error) {
			*called = true
			return next.Do(req)
		})
	}
}

func TestClient_CallMiddleware(t *testing.T) {
	var called bool
	m := newClientMiddleware(&called)

	c := fdhttp.NewClient()
	c.Use(m)

	var srvCalled bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srvCalled = true
	}))
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, bytes.NewBufferString(""))
	resp, err := c.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.True(t, called)
	assert.True(t, srvCalled)
}