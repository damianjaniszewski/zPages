package zpages

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	appName    = "zpages"
	appVersion = "v0.0.4"
)

const (
	ServiceTypeInfrastructure = "infrastructure"
	ServiceTypePlatform       = "platform"
	ServiceTypeApplication    = "application"
	ServiceTypeExternal       = "external"
)

const (
	HealthStatusOK           = "ok"
	HealthStatusFailed       = "failed"
	HealthStatusUnknown      = "unknown"
	HealthStatusOKForced     = "ok-forced"
	HealthStatusFailedForced = "failed-forced"

	ReadinessStatusReady          = "ready"
	ReadinessStatusNotReady       = "not-ready"
	ReadinessStatusUnknown        = "unknown"
	ReadinessStatusReadyForced    = "ready-forced"
	ReadinessStatusNotReadyForced = "not-ready-forced"
)

const (
	LogFormatText = "text"
	LogFormatJSON = "json"
)

// Handler
type Handler struct {
	Version       Version
	ServiceStatus ServiceStatus
	LogLevel      LogLevel

	ShutdownChannel chan os.Signal

	Logger *logrus.Logger
}

func (z *Handler) Init(appName string, appVersion string, instanceGUID string, instanceAddress string, appType string,
	logLevel string, debugLevel int, logAs string) {

	z.ServiceStatus = ServiceStatus{
		GUID: instanceGUID, Name: appName, Type: appType, URI: instanceAddress,
		Health: HealthStatus{Status: HealthStatusOK}, Readiness: ReadinessStatus{Status: ReadinessStatusReady},
		Updated: time.Now(),
	}
	z.LogLevel = LogLevel{logLevel, debugLevel, logAs}

	z.Version.Dependencies = append(z.Version.Dependencies, Version{Module: appName, Version: appVersion})

	z.ServiceStatus.mutexHealth.Lock()
	defer z.ServiceStatus.mutexHealth.Unlock()
	z.ServiceStatus.statusHealth = true
	z.ServiceStatus.statusHealthyForced = false
	z.ServiceStatus.statusNotHealthyForced = false

	z.ServiceStatus.mutexReady.Lock()
	defer z.ServiceStatus.mutexReady.Unlock()
	z.ServiceStatus.statusReady = true
	z.ServiceStatus.statusReadyForced = false
	z.ServiceStatus.statusNotReadyForced = false

	z.Logger.WithFields(logrus.Fields{
		"statusHealth": z.ServiceStatus.statusHealth, "statusNotHealthyForced": z.ServiceStatus.statusNotHealthyForced,
		"statusReady": z.ServiceStatus.statusReady, "statusNotReadyForced": z.ServiceStatus.statusNotReadyForced,
		"logLevel": z.LogLevel.Log, "debugLevel": z.LogLevel.Debug, "logAs": z.LogLevel.Format,
	}).Infof("%s %s initialized", appName, appVersion)
}
