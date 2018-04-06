package fdhttp

import (
	"context"
	"net/http"
	"time"
)

// HealthCheckService is the interface that your service need to provide to
// be able to health check
type HealthCheckService interface {
	HealthCheck() (interface{}, error)
}

// HealthCheckResponse is the main json response from healthcheck endpoint
type HealthCheckResponse struct {
	Status  bool `json:"status"`
	Version struct {
		Tag    string `json:"tag"`
		Commit string `json:"commit"`
	} `json:"version"`
	Elapsed time.Duration                         `json:"elapsed"`
	Checks  map[string]HealthCheckServiceResponse `json:"checks,omitempty"`
}

// HealthCheckServiceResponse is the return of each service that can provide
// details
type HealthCheckServiceResponse struct {
	Status  bool          `json:"status"`
	Elapsed time.Duration `json:"elapsed"`
	Detail  interface{}   `json:"detail,omitempty"`
	Error   interface{}   `json:"error,omitempty"`
}

// HealthCheckServiceError can be returned as a error and Detail()
// will populate Error response
type HealthCheckServiceError interface {
	error
	Detail() interface{}
}

var _ Handler = &HealchCheckHandler{}

// DefaultHealthCheckURL is the urlto access your health check.
// Note you can specify a Prefix (inside of HealchCheckHandler object)
// and you also can check a specific service using: Prefix + DefaultHealthCheckURL + "/<service-name>"
var DefaultHealthCheckURL = "/health/check"

// HealchCheckHandler is a valid http handler to export app version and it also checks
// service registred.
type HealchCheckHandler struct {
	// Prefix will be prefix the fdhttp.DefaultHealthCheckURL.
	Prefix string

	tag      string
	commit   string
	services map[string]HealthCheckService
}

// NewHealthCheckHandler create a new healthcheck handler
func NewHealthCheckHandler(tag, commit string) *HealchCheckHandler {
	return &HealchCheckHandler{
		tag:      tag,
		commit:   commit,
		services: make(map[string]HealthCheckService),
	}
}

func (h *HealchCheckHandler) Init(r *Router) {
	r.GET(h.Prefix+DefaultHealthCheckURL, h.Get)
	r.GET(h.Prefix+DefaultHealthCheckURL+"/:service", h.Get)
}

// Register a new healthcheck Service
func (h *HealchCheckHandler) Register(name string, s HealthCheckService) {
	h.services[name] = s
}

func (h *HealchCheckHandler) Get(ctx context.Context) (int, interface{}, error) {
	started := time.Now()

	serviceParam := RouteParam(ctx, "service")

	statusCode := http.StatusOK
	resp := HealthCheckResponse{
		Status: true,
		Checks: make(map[string]HealthCheckServiceResponse),
	}
	resp.Version.Tag = h.tag
	resp.Version.Commit = h.commit

	for name, s := range h.services {
		if serviceParam != "" && name != serviceParam {
			continue
		}

		started := time.Now()
		serviceCheck := HealthCheckServiceResponse{
			Status: true,
		}

		detail, err := s.HealthCheck()
		if err != nil {
			statusCode = http.StatusServiceUnavailable
			resp.Status = false
			serviceCheck.Status = false

			if hcErr, ok := err.(HealthCheckServiceError); ok {
				serviceCheck.Error = hcErr.Detail()
			} else {
				serviceCheck.Error = err.Error()
			}

		} else {
			serviceCheck.Detail = detail
		}

		serviceCheck.Elapsed = time.Since(started) / time.Millisecond
		resp.Checks[name] = serviceCheck
	}

	resp.Elapsed = time.Since(started) / time.Millisecond

	return statusCode, resp, nil
}
