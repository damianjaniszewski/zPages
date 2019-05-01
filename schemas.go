package zpages

import (
	"sync"
	"time"
)

type Info struct {
	Version  Version       `json:"version"`
	Status   ServiceStatus `json:"status"`
	Env      Env           `json:"env"`
	Loglevel LogLevel      `json:"logLevel"`
}

type Version struct {
	Module       string    `json:"module"`
	Version      string    `json:"version"`
	Dependencies []Version `json:"dependencies,omitempty"`
}

type ServiceStatus struct {
	GUID             string          `json:"guid"`
	Name             string          `json:"name"`
	Type             string          `json:"type"` // [infrastructure, platform, application, external]
	Health           HealthStatus    `json:"healthStatus"`
	Readiness        ReadinessStatus `json:"readinessStatus"`
	Updated          time.Time       `json:"updated"`
	URI              string          `json:"uri"`
	UpstreamServices []ServiceStatus `json:"upstreamServices,omitempty"`

	mutexHealth sync.Mutex
	mutexReady  sync.Mutex

	// healthProbe

	statusHealth           bool
	statusHealthyForced    bool
	statusNotHealthyForced bool

	// readinessProbe

	statusReady          bool
	statusReadyForced    bool
	statusNotReadyForced bool
}

type HealthStatus struct {
	Status string `json:"status"` // [ok, failed, unknown, ok-forced, failed-forced]
}

type ReadinessStatus struct {
	Status string `json:"status"` // [ready, not-ready, unknown, ready-forced, not-ready-forced]
}

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type LogLevel struct {
	Log    string `json:"log"` // [PANIC, FATAL, ERROR, WARNING, INFO, DEBUG, TRACE]
	Debug  int    `json:"debug"`
	Format string `json:"format"` // [text, json]
}

type Error struct {
	Message            string   `json:"message"`
	Details            string   `json:"details,omitempty"`
	RecommendedActions []string `json:"recommendedActions,omitempty"`
	NestedErrors       string   `json:"nestedErrors,omitempty"`
	ErrorSource        string   `json:"errorSource"`
	ErrorCode          string   `json:"errorCode"`
	Data               string   `json:"data,omitempty"`
}
