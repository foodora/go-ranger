package fdmiddleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/foodora/go-ranger/fdhttp"
	"github.com/foodora/go-ranger/fdhttp/fdmiddleware"
	"github.com/stretchr/testify/assert"
)

type dummyLog struct {
	PrintfMsg string
}

func (l *dummyLog) Printf(format string, v ...interface{}) {
	l.PrintfMsg += fmt.Sprintf(format, v...)
}

func TestNewLogMiddleware(t *testing.T) {
	logger := &dummyLog{}
	logMiddleware := fdmiddleware.NewLogMiddleware()
	logMiddleware.SetLogger(logger)

	called := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusBadRequest)
	}

	ts := httptest.NewServer(logMiddleware.Wrap(http.HandlerFunc(handler)))
	defer ts.Close()

	http.Get(ts.URL + "/foo")

	assert.True(t, called)
	assert.Regexp(t, "^127.0.0.1 \\[([0-9]+\\.)?[0-9]+[nµm]?s\\] \"GET /foo HTTP/1.1\" 400 Bad Request \"Go-http-client/1.1\"$", logger.PrintfMsg)
}

func TestNewLogMiddleware_CaptureStatusCodeSentToOriginalResponse(t *testing.T) {
	// It's not clear if we should solve this problem. The problem here is if
	// the user call fdhttp.Response(ctx) will be the original http.ResponseWriter
	// ignoring any middleware that inject something different, like logger and newrelic, e.g.
	// At the same point if we force to override it, we don't have any way to bypass
	// these middleware, which is one of our usecases. There're a situation where
	// we want to return failure to clients but ignore newrelic alerts.

	fdmiddleware.RequestLogFormat = "{{.Response.StatusCode}}"

	logger := &dummyLog{}
	logMiddleware := fdmiddleware.NewLogMiddleware()
	logMiddleware.SetLogger(logger)

	called := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		called = true
		originalW := fdhttp.Response(req.Context())
		originalW.WriteHeader(http.StatusBadRequest)
	}

	router := fdhttp.NewRouter()
	router.Use(logMiddleware)
	router.StdGET("/foo", handler)

	ts := httptest.NewServer(router)
	defer ts.Close()

	http.Get(ts.URL + "/foo")

	assert.True(t, called)
	// assert.Equal(t, "400", logger.PrintfMsg)
}

func TestNewLogMiddleware_DifferentLogFormat(t *testing.T) {
	logger := &dummyLog{}
	fdmiddleware.RequestLogFormat = "{{.Method}} {{.RequestURI}} {{.Response.StatusCode}} {{.Response.StatusText}}"

	logMiddleware := fdmiddleware.NewLogMiddleware()
	logMiddleware.SetLogger(logger)

	called := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusBadRequest)
	}

	ts := httptest.NewServer(logMiddleware.Wrap(http.HandlerFunc(handler)))
	defer ts.Close()

	http.Get(ts.URL + "/foo")

	assert.True(t, called)
	assert.Equal(t, "GET /foo 400 Bad Request", logger.PrintfMsg)
}

func TestNewLogMiddleware_CallFuncInEachRequest(t *testing.T) {
	logger := &dummyLog{}

	logMiddleware := fdmiddleware.NewLogMiddleware()
	logMiddleware.SetLoggerFunc(func(logReq *fdmiddleware.LogRequest) {
		logger.Printf("%s %s %d", logReq.Method, logReq.RequestURI, logReq.Response.StatusCode)
	})

	called := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}

	ts := httptest.NewServer(logMiddleware.Wrap(http.HandlerFunc(handler)))
	defer ts.Close()

	http.Get(ts.URL + "/foo")

	assert.True(t, called)
	assert.Equal(t, "GET /foo 200", logger.PrintfMsg)
}

func TestNewLogMiddleware_CanGetError(t *testing.T) {
	logger := &dummyLog{}
	logMiddleware := fdmiddleware.NewLogMiddleware()
	logMiddleware.SetLoggerFunc(func(logReq *fdmiddleware.LogRequest) {
		err := fdhttp.ResponseError(logReq.Context())
		assert.IsType(t, &fdhttp.Error{}, err)
		logger.Printf("%s", err)
	})

	handlerErr := &fdhttp.Error{
		Code:    "my_error",
		Message: "details",
	}

	router := fdhttp.NewRouter()
	router.Use(logMiddleware)
	router.GET("/foo", func(ctx context.Context) (int, interface{}) {
		return http.StatusBadRequest, handlerErr
	})
	router.Init()

	ts := httptest.NewServer(router)
	defer ts.Close()

	http.Get(ts.URL + "/foo")

	assert.Equal(t, handlerErr.Error(), logger.PrintfMsg)
}
