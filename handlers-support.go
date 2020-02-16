package zpages

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func (z *Handler) SupportQuiesce(w http.ResponseWriter, r *http.Request) {
	log := z.Logger.WithFields(logrus.Fields{"app": z.Version.Module, "version": z.Version.Version, "instance": z.ServiceStatus.URI, "guid": z.ServiceStatus.GUID})

	z.responseHeader(w)

	w.WriteHeader(http.StatusOK)

	z.SetNotReadyForced()

	out, _ := json.Marshal(z.ServiceStatus)
	fmt.Fprintf(w, string(out))

	log.WithFields(logrus.Fields{
		"statusReady": z.ServiceStatus.statusReady, "statusNotReadyForced": z.ServiceStatus.statusNotReadyForced,
	}).Warnf("%s SupportQuiesce: set forced not ready", appName)

}

func (z *Handler) SupportResume(w http.ResponseWriter, r *http.Request) {
	log := z.Logger.WithFields(logrus.Fields{"app": z.Version.Module, "version": z.Version.Version, "instance": z.ServiceStatus.URI, "guid": z.ServiceStatus.GUID})

	z.responseHeader(w)

	w.WriteHeader(http.StatusOK)

	z.UnsetNotReadyForced()

	out, _ := json.Marshal(z.ServiceStatus)
	fmt.Fprintf(w, string(out))

	log.WithFields(logrus.Fields{
		"statusReady": z.ServiceStatus.statusReady, "statusNotReadyForced": z.ServiceStatus.statusNotReadyForced,
	}).Warnf("%s SupportResume: unset forced not ready", appName)
}

func (z *Handler) SupportFail(w http.ResponseWriter, r *http.Request) {
	log := z.Logger.WithFields(logrus.Fields{"app": z.Version.Module, "version": z.Version.Version, "instance": z.ServiceStatus.URI, "guid": z.ServiceStatus.GUID})

	z.responseHeader(w)

	w.WriteHeader(http.StatusOK)

	z.SetUnhealthyForced()

	out, _ := json.Marshal(z.ServiceStatus)
	fmt.Fprintf(w, string(out))

	log.WithFields(logrus.Fields{
		"statusHealth": z.ServiceStatus.statusHealth, "statusNotHealthyForced": z.ServiceStatus.statusNotHealthyForced,
	}).Warnf("%s SupportFail: set forced unhealthy", appName)
}

func (z *Handler) SupportQuit(w http.ResponseWriter, r *http.Request) {
	log := z.Logger.WithFields(logrus.Fields{"app": z.Version.Module, "version": z.Version.Version, "instance": z.ServiceStatus.URI, "guid": z.ServiceStatus.GUID})

	z.responseHeader(w)

	w.WriteHeader(http.StatusOK)

	z.SetNotReadyForced()
	z.ShutdownChannel <- os.Interrupt
	z.SetUnhealthyForced()

	out, _ := json.Marshal(z.ServiceStatus)
	fmt.Fprintf(w, string(out))

	log.WithFields(logrus.Fields{
		"statusHealth": z.ServiceStatus.statusHealth, "statusNotHealthyForced": z.ServiceStatus.statusNotHealthyForced,
	}).Warnf("%s SupportQuit: set forced not ready and forced unhealthy", appName)
}

func (z *Handler) SupportRestart(w http.ResponseWriter, r *http.Request) {
	log := z.Logger.WithFields(logrus.Fields{"app": z.Version.Module, "version": z.Version.Version, "instance": z.ServiceStatus.URI, "guid": z.ServiceStatus.GUID})

	z.responseHeader(w)

	w.WriteHeader(http.StatusOK)

	z.SetNotReadyForced()

	log.WithFields(logrus.Fields{
		"statusReady": z.ServiceStatus.statusReady, "statusNotReadyForced": z.ServiceStatus.statusNotReadyForced,
	}).Warnf("%s SupportRestart: restarting...", appName)

	z.UnsetNotReadyForced()

	log.WithFields(logrus.Fields{
		"statusReady": z.ServiceStatus.statusReady, "statusNotReadyForced": z.ServiceStatus.statusNotReadyForced,
	}).Warnf("%s SupportRestart: restarted", appName)

	out, _ := json.Marshal(z.ServiceStatus)
	fmt.Fprintf(w, string(out))
}

func (z *Handler) SupportCrash(w http.ResponseWriter, r *http.Request) {
	log := z.Logger.WithFields(logrus.Fields{"app": z.Version.Module, "version": z.Version.Version, "instance": z.ServiceStatus.URI, "guid": z.ServiceStatus.GUID})

	z.responseHeader(w)

	w.WriteHeader(http.StatusOK)
	out, _ := json.Marshal(z.ServiceStatus)
	fmt.Fprintf(w, string(out))

	log.Fatalf("%s SupportCrash: crashing...", appName)
}

func (z *Handler) SetNotReady() {
	z.ServiceStatus.mutexReady.Lock()
	defer z.ServiceStatus.mutexReady.Unlock()

	z.ServiceStatus.Readiness.Status = ReadinessStatusNotReady

	z.ServiceStatus.statusReady = false
	z.ServiceStatus.statusNotReadyForced = false
	z.ServiceStatus.statusReadyForced = false
}

func (z *Handler) UnsetNotReady() {
	z.ServiceStatus.mutexReady.Lock()
	defer z.ServiceStatus.mutexReady.Unlock()

	z.ServiceStatus.Readiness.Status = ReadinessStatusReady

	z.ServiceStatus.statusReady = true
	z.ServiceStatus.statusNotReadyForced = false
	z.ServiceStatus.statusReadyForced = false
}

func (z *Handler) SetNotReadyForced() {
	z.ServiceStatus.mutexReady.Lock()
	defer z.ServiceStatus.mutexReady.Unlock()

	z.ServiceStatus.Readiness.Status = ReadinessStatusNotReadyForced

	z.ServiceStatus.statusNotReadyForced = true
	z.ServiceStatus.statusReadyForced = false
}

func (z *Handler) UnsetNotReadyForced() {
	z.ServiceStatus.mutexReady.Lock()
	defer z.ServiceStatus.mutexReady.Unlock()

	z.ServiceStatus.Readiness.Status = ReadinessStatusReady

	z.ServiceStatus.statusNotReadyForced = false
	z.ServiceStatus.statusReadyForced = false
}

func (z *Handler) SetUnhealthyForced() {
	z.ServiceStatus.mutexHealth.Lock()
	defer z.ServiceStatus.mutexHealth.Unlock()

	z.ServiceStatus.Health.Status = HealthStatusFailedForced

	z.ServiceStatus.statusNotHealthyForced = true
	z.ServiceStatus.statusHealthyForced = false
}

func (z *Handler) SupportEnv(w http.ResponseWriter, r *http.Request) {
	log := z.Logger.WithFields(logrus.Fields{"app": z.Version.Module, "version": z.Version.Version, "instance": z.ServiceStatus.URI, "guid": z.ServiceStatus.GUID})

	z.responseHeader(w)

	log.Debugf("%s SupportEnv", appName)

	if r.Method == "GET" {
		env := make([]Env, 0)

		for _, item := range os.Environ() {
			variable := strings.Split(item, "=")
			env = append(env, Env{variable[0], variable[1]})
		}

		if out, err := json.Marshal(env); err != nil {
			log.Errorf("%s SupportEnv: %v", appName, err)
			w.WriteHeader(http.StatusInternalServerError)

			e := &Error{
				Message:     err.Error(),
				ErrorSource: appName,
			}
			out, _ = json.Marshal(e)
			fmt.Fprintf(w, string(out))
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, string(out))
		}
	}
}

func (z *Handler) SupportVersion(w http.ResponseWriter, r *http.Request) {
	log := z.Logger.WithFields(logrus.Fields{"app": z.Version.Module, "version": z.Version.Version, "instance": z.ServiceStatus.URI, "guid": z.ServiceStatus.GUID})

	z.responseHeader(w)

	log.Debugf("%s SupportVersion: %+v", appName, z.Version)

	if r.Method == "GET" {

		if out, err := json.Marshal(z.Version); err != nil {
			log.Errorf("%s SupportEnv: %v", appName, err)
			w.WriteHeader(http.StatusInternalServerError)

			e := &Error{
				Message:     err.Error(),
				ErrorSource: appName,
			}
			out, _ = json.Marshal(e)
			fmt.Fprintf(w, string(out))
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, string(out))
		}
	}
}

func (z *Handler) SupportLogLevel(w http.ResponseWriter, r *http.Request) {
	log := z.Logger.WithFields(logrus.Fields{"app": z.Version.Module, "version": z.Version.Version, "instance": z.ServiceStatus.URI, "guid": z.ServiceStatus.GUID})

	z.responseHeader(w)

	if r.Method == "GET" {
		log.WithFields(logrus.Fields{
			"logLevel": z.LogLevel.Log, "debugLevel": z.LogLevel.Debug, "logAs": z.LogLevel.Format,
		}).Debugf("%s SupportLogLevel", appName)

		if out, err := json.Marshal(z.LogLevel); err != nil {
			log.Errorf("%s SupportLogLevel: %v", appName, err)
			w.WriteHeader(http.StatusInternalServerError)

			e := &Error{
				Message:     err.Error(),
				ErrorSource: appName,
			}
			out, _ = json.Marshal(e)
			fmt.Fprintf(w, string(out))
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, string(out))
		}
	}

	if r.Method == "PUT" {
		var data bytes.Buffer
		if _, err := data.ReadFrom(r.Body); err != nil {
			log.Errorf("%s SupportLogLevel: %v", appName, err)
			w.WriteHeader(http.StatusNoContent)

			e := &Error{
				Message:     err.Error(),
				ErrorSource: appName,
			}
			out, _ := json.Marshal(e)
			fmt.Fprintf(w, string(out))
		} else {
			var loglevel LogLevel
			if err := json.Unmarshal(data.Bytes(), &loglevel); err != nil {
				log.Errorf("%s SupportLogLevel: %v", appName, err)

				e := &Error{
					Message:     err.Error(),
					ErrorSource: appName,
				}
				out, _ := json.Marshal(e)
				fmt.Fprintf(w, string(out))
			} else {
				log.WithFields(logrus.Fields{
					"logLevel": loglevel.Log, "debugLevel": loglevel.Debug, "logAs": loglevel.Format,
				}).Debugf("%s SupportLogLevel", appName)

				z.setLogLevel(loglevel.Log, loglevel.Debug, loglevel.Format)

				if out, err := json.Marshal(z.LogLevel); err != nil {
					log.Errorf("%s SupportLogLevel: %v", appName, err)
					w.WriteHeader(http.StatusInternalServerError)

					e := &Error{
						Message:     err.Error(),
						ErrorSource: appName,
					}
					out, _ = json.Marshal(e)
					fmt.Fprintf(w, string(out))
				} else {
					w.WriteHeader(http.StatusOK)
					fmt.Fprintf(w, string(out))
				}
			}
		}
	}
}

func (z *Handler) setLogLevel(logLevel string, debugLevel int, logAs string) {
	switch logLevel {
	case "PANIC":
		z.Logger.Level = logrus.PanicLevel
	case "FATAL":
		z.Logger.Level = logrus.FatalLevel
	case "ERROR":
		z.Logger.Level = logrus.ErrorLevel
	case "WARNING":
		z.Logger.Level = logrus.WarnLevel
	case "INFO":
		z.Logger.Level = logrus.InfoLevel
	case "DEBUG":
		z.Logger.Level = logrus.DebugLevel
	case "TRACE":
		z.Logger.Level = logrus.TraceLevel
	default:
		z.Logger.Level = logrus.InfoLevel
	}
	z.LogLevel.Log = logLevel

	z.LogLevel.Debug = debugLevel

	logFullTimestamp := false
	if debugLevel > 0 {
		logFullTimestamp = true
	}

	switch logAs {
	case "text":
		z.Logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: logFullTimestamp})
	case "json":
		z.Logger.SetFormatter(&logrus.JSONFormatter{})
	default:
		z.Logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: logFullTimestamp})
	}
	z.LogLevel.Format = logAs
}

func (z *Handler) responseHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
}
