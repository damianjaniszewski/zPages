package zpages

import (
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

var (
	debugLevel = 0
	log        = logrus.New()
)

// Health status
type Health struct {
	mutexHealth sync.Mutex
	mutexReady  sync.Mutex

	statusHealth           bool
	statusHealthyForced    bool
	statusNotHealthyForced bool

	statusReady          bool
	statusReadyForced    bool
	statusNotReadyForced bool

	chQuit chan os.Signal

	appName         string
	instanceAddress string
	instanceGUID    uuid.UUID
}

func New(chQuit *chan os.Signal, appName string, instanceAddress string, instanceGUID uuid.UUID) *Health {
	h := new(Health)

	h.statusHealth = true
	h.statusHealthyForced = false
	h.statusNotHealthyForced = false
	h.statusReady = true
	h.statusReadyForced = false
	h.statusNotReadyForced = false

	h.chQuit = *chQuit

	h.appName = appName
	h.instanceAddress = instanceAddress
	h.instanceGUID = instanceGUID

	return h
}

func (h *Health) SupportQuiesce(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.WriteHeader(http.StatusOK)

	log.WithFields(logrus.Fields{
		"app": h.appName, "instance": h.instanceAddress, "guid": h.instanceGUID,
	}).Warn("got /quiesce request, setting service to forced not ready")

	h.setNotReady()
}

func (h *Health) SupportResume(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.WriteHeader(http.StatusOK)

	log.WithFields(logrus.Fields{
		"app": h.appName, "instance": h.instanceAddress, "guid": h.instanceGUID,
	}).Warn("got /resume request, unsetting forced not ready")

	h.unsetNotReady()
}

func (h *Health) SupportFail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.WriteHeader(http.StatusOK)

	log.WithFields(logrus.Fields{
		"app": h.appName, "instance": h.instanceAddress, "guid": h.instanceGUID,
	}).Warn("got /fail request, setting service to forced unhealthy")

	h.setUnhealthy()
}

func (h *Health) SupportQuit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.WriteHeader(http.StatusOK)

	log.WithFields(logrus.Fields{
		"app": h.appName, "instance": h.instanceAddress, "guid": h.instanceGUID,
	}).Warn("got /quit request, setting service to forced not ready and to forced unhealthy")

	h.setNotReady()
	h.chQuit <- os.Interrupt
	h.setUnhealthy()
}

func (h *Health) SupportRestart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.WriteHeader(http.StatusOK)

	log.WithFields(logrus.Fields{
		"app": h.appName, "instance": h.instanceAddress, "guid": h.instanceGUID,
	}).Warn("got /restart request, restarting...")
}

func (h *Health) SupportCrash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.WriteHeader(http.StatusInternalServerError)

	log.WithFields(logrus.Fields{
		"app": h.appName, "instance": h.instanceAddress, "guid": h.instanceGUID,
	}).Fatal("got /crash request, crashing...")
}

func (h *Health) setNotReady() {
	h.mutexReady.Lock()
	defer h.mutexReady.Unlock()

	h.statusNotReadyForced = true
	h.statusReadyForced = false
}

func (h *Health) unsetNotReady() {
	h.mutexReady.Lock()
	defer h.mutexReady.Unlock()

	h.statusNotReadyForced = false
	h.statusReadyForced = false
}

func (h *Health) setUnhealthy() {
	h.mutexHealth.Lock()
	defer h.mutexHealth.Unlock()

	h.statusNotHealthyForced = true
	h.statusHealthyForced = false
}

func (h *Health) Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")

	h.mutexHealth.Lock()
	defer h.mutexHealth.Unlock()

	log.WithFields(logrus.Fields{
		"app": h.appName, "instance": h.instanceAddress, "guid": h.instanceGUID,
	}).Debugf("healthz status: %v, forced not healthy: %v", h.statusHealth, h.statusNotHealthyForced)

	if h.statusNotHealthyForced == true {
		// 501
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	if h.statusHealthyForced == true {
		// 202
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		return
	}

	if h.statusHealth == true {
		// 200
		w.WriteHeader(http.StatusOK)
	} else {
		// 503
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func (h *Health) Readyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")

	h.mutexReady.Lock()
	defer h.mutexReady.Unlock()

	log.WithFields(logrus.Fields{
		"app": h.appName, "instance": h.instanceAddress, "guid": h.instanceGUID,
	}).Debugf("readyz status: %v, forced not ready: %v", h.statusReady, h.statusNotReadyForced)

	if h.statusNotReadyForced == true {
		// 501
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	if h.statusReadyForced == true {
		// 203
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		return
	}

	if h.statusReady == true {
		// 200
		w.WriteHeader(http.StatusOK)
	} else {
		// 503
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func (h *Health) SetLogLevel(logLevel string, debug int, logAs string) {
	switch logLevel {
	case "PANIC":
		log.Level = logrus.PanicLevel
	case "FATAL":
		log.Level = logrus.FatalLevel
	case "ERROR":
		log.Level = logrus.ErrorLevel
	case "WARNING":
		log.Level = logrus.WarnLevel
	case "INFO":
		log.Level = logrus.InfoLevel
	case "DEBUG":
		log.Level = logrus.DebugLevel
	case "TRACE":
		log.Level = logrus.TraceLevel
	default:
		log.Level = logrus.InfoLevel
	}

	switch logAs {
	case "text":
		log.SetFormatter(&logrus.TextFormatter{})
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{})
	default:
		log.SetFormatter(&logrus.TextFormatter{})
	}

	debugLevel = debug

	if debugLevel > 2 {
		log.SetReportCaller(true)
		switch logAs {
		case "text":
			log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
		case "json":
			log.SetFormatter(&logrus.JSONFormatter{})
		default:
			log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
		}
	}
}

// ServiceStatus:
//       description: Status of the service and all upstream services
//       type: object
//       required: [id, name, type, health, readiness, updated]
//       properties:
//         id:
//           type: string
//           format: uuid
//           example: '0729a580-2240-11e6-9eb5-0002a5d5c51b'
//         name:
//           type: string
//           example: 'data-store'
//         type:
//           enum: [infrastructure, platform, application, external]
//           example: platform
//           description: Type of the service.
//         health:
//           $ref: '#/components/schemas/HealthStatus'
//         readiness:
//           $ref: '#/components/schemas/ReadinessStatus'
//         updated:
//           type: string
//           format: date-time
//           # date-time format as described in RFC 3339, section 5.6 (https://tools.ietf.org/html/rfc3339#section-5.6)
//           example: '2019-07-18T17:32:28-02:00'
//           description: Date and time of last health and/or readiness update.
//         uri:
//           type: string
//           format: uri
//           example: 'http://data-store.smpaas-platform-svcs.svc.cluster.local:8080'
//         upstreamServices:
//           type: array
//           items:
//             $ref: '#/components/schemas/ServiceStatus'

//     HealthStatus:
//       description: Health status of the service
//       type: object
//       required: [status]
//       properties:
//         status:
//           type: string
//           enum: [ok, failed, unknown, ok-forced, failed-forced]
//           example: ok

//     ReadinessStatus:
//       description: Readiness status of the service
//       type: object
//       required: [status]
//       properties:
//         status:
//           type: string
//           enum: [ready, not-ready, unknown, not-ready-forced]
// 					example: ok

func init() {
	level, _ := os.LookupEnv("LOGLEVEL")
	switch level {
	case "PANIC":
		log.Level = logrus.PanicLevel
	case "FATAL":
		log.Level = logrus.FatalLevel
	case "ERROR":
		log.Level = logrus.ErrorLevel
	case "WARNING":
		log.Level = logrus.WarnLevel
	case "INFO":
		log.Level = logrus.InfoLevel
	case "DEBUG":
		log.Level = logrus.DebugLevel
	case "TRACE":
		log.Level = logrus.TraceLevel
	default:
		log.Level = logrus.InfoLevel
	}

	logAs, _ := os.LookupEnv("LOGAS")
	switch logAs {
	case "text":
		log.SetFormatter(&logrus.TextFormatter{})
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{})
	default:
		log.SetFormatter(&logrus.TextFormatter{})
	}

	debug, debugExists := os.LookupEnv("DEBUGLEVEL")
	if debugExists {
		debugLevel, err := strconv.Atoi(debug)
		if err != nil {
			log.WithFields(logrus.Fields{}).Errorf("error converting DEBUGLEVEL %s: %v", debug, err)
			debugLevel = 0
		}

		if debugLevel > 2 {
			log.SetReportCaller(true)
			switch logAs {
			case "text":
				log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
			case "json":
				log.SetFormatter(&logrus.JSONFormatter{})
			default:
				log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
			}
		}
	}

	log.Out = os.Stdout

	log.WithFields(logrus.Fields{
		"logLevel": log.Level, "debugLevel": debugLevel,
	}).Info("zpages initialized")
}
