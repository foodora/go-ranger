package fdhttp_test

import (
	"context"
	"log"
	"net/http"

	"github.com/foodora/go-ranger/fdhttp"
	"github.com/foodora/go-ranger/fdhttp/fdmiddleware"
)

func testMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s", req.Method, req.URL.String())
		next.ServeHTTP(w, req)
	})
}

type testHandler struct{}

func (h *testHandler) Init(router *fdhttp.Router) {
	router.PUT("/entity/:id", h.PutHandler)
}

func (h *testHandler) PutHandler(ctx context.Context) (int, interface{}) {
	id := fdhttp.RouteParam(ctx, "id")

	var bodyReq map[string]interface{}

	err := fdhttp.RequestBodyJSON(ctx, &bodyReq)
	if err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, map[string]interface{}{
		"id":   id,
		"body": bodyReq,
	}
}

func Example() {
	srv := fdhttp.NewServer("8080")

	testMw := fdmiddleware.MiddlewareFunc(testMiddleware)

	router := fdhttp.NewRouter()
	router.Use(testMw)
	router.Register(&testHandler{})

	err := srv.Start(router)
	if err != nil {
		log.Fatalln("Cannot run server:", err)
	}
}
