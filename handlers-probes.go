package zpages

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func (z *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	log := z.Logger.WithFields(logrus.Fields{"app": z.Version.Module, "version": z.Version.Version, "instance": z.ServiceStatus.URI, "guid": z.ServiceStatus.GUID})

	z.responseHeader(w)

	z.ServiceStatus.mutexHealth.Lock()
	defer z.ServiceStatus.mutexHealth.Unlock()

	if z.LogLevel.Debug >= 4 {
		log.WithFields(logrus.Fields{
			"statusHealth": z.ServiceStatus.statusHealth, "statusNotHealthyForced": z.ServiceStatus.statusNotHealthyForced,
		}).Debugf("%s Healthz: healthy: %v, forced not healthy: %v", appName, z.ServiceStatus.statusHealth, z.ServiceStatus.statusNotHealthyForced)
	}

	if z.ServiceStatus.statusNotHealthyForced == true {
		// 501
		w.WriteHeader(http.StatusNotImplemented)
	} else if z.ServiceStatus.statusHealthyForced == true {
		// 202
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
	} else if z.ServiceStatus.statusHealth == true {
		// 200
		w.WriteHeader(http.StatusOK)
	} else {
		// 503
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	status, _ := json.Marshal(z.ServiceStatus)
	fmt.Fprintf(w, string(status))
}

func (z *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	log := z.Logger.WithFields(logrus.Fields{
		"app": z.Version.Module, "version": z.Version.Version, "instance": z.ServiceStatus.URI, "guid": z.ServiceStatus.GUID,
	})

	z.responseHeader(w)

	z.ServiceStatus.mutexReady.Lock()
	defer z.ServiceStatus.mutexReady.Unlock()

	if z.LogLevel.Debug >= 4 {
		log.WithFields(logrus.Fields{
			"statusReady": z.ServiceStatus.statusReady, "statusNotReadyForced": z.ServiceStatus.statusNotReadyForced,
		}).Debugf("%s Readyz: ready: %v, forced not ready: %v", appName, z.ServiceStatus.statusReady, z.ServiceStatus.statusNotReadyForced)
	}

	if z.ServiceStatus.statusNotReadyForced {
		// 501
		w.WriteHeader(http.StatusNotImplemented)
	} else if z.ServiceStatus.statusReadyForced {
		// 203
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
	} else if z.ServiceStatus.statusReady {
		// 200
		w.WriteHeader(http.StatusOK)
	} else {
		// 503
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	status, _ := json.Marshal(z.ServiceStatus)
	fmt.Fprintf(w, string(status))
}
