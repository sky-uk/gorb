package core

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "gorb" // For Prometheus metrics.
)

var (
	serviceHealth = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "service_health",
		Help:      "Health of the load balancer service",
	}, []string{"name", "host", "port", "protocol"})

	serviceBackends = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "service_backends",
		Help:      "Number of backends in the load balancer service",
	}, []string{"name", "host", "port", "protocol"})

	serviceBackendUptimeTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "service_backend_uptime_seconds",
		Help:      "Uptime in seconds of a backend service",
	}, []string{"service_name", "name", "host", "port"})

	serviceBackendHealth = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "service_backend_health",
		Help:      "Health of a backend service",
	}, []string{"service_name", "name", "host", "port"})

	serviceBackendStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "service_backend_status",
		Help:      "Status of a backend service",
	}, []string{"service_name", "name", "host", "port"})

	serviceBackendWeight = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "service_backend_weight",
		Help:      "Weight of a backend service",
	}, []string{"service_name", "name", "host", "port"})
)

type Exporter struct {
	ctx *Context
}

func NewExporter(ctx *Context) *Exporter {
	return &Exporter{
		ctx: ctx,
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	serviceHealth.Describe(ch)
	serviceBackends.Describe(ch)
	serviceBackendUptimeTotal.Describe(ch)
	serviceBackendHealth.Describe(ch)
	serviceBackendStatus.Describe(ch)
	serviceBackendWeight.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	if err := e.collect(); err != nil {
		log.Errorf("error collecting metrics: %s", err)
		return
	}
	serviceHealth.Collect(ch)
	serviceBackends.Collect(ch)
	serviceBackendUptimeTotal.Collect(ch)
	serviceBackendHealth.Collect(ch)
	serviceBackendStatus.Collect(ch)
	serviceBackendWeight.Collect(ch)
}

func (e *Exporter) collect() error {
	e.ctx.mutex.RLock()
	defer e.ctx.mutex.RUnlock()

	for serviceName := range e.ctx.services {
		service, err := e.ctx.GetService(serviceName)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error getting service: %s", serviceName))
		}

		serviceHealth.WithLabelValues(serviceName, service.Options.VIP.String(), fmt.Sprintf("%d", service.Options.Port),
			service.Options.Protocol).
			Set(service.Health)

		serviceBackends.WithLabelValues(serviceName, service.Options.VIP.String(), fmt.Sprintf("%d", service.Options.Port),
			service.Options.Protocol).
			Set(float64(len(service.Backends)))

		for i := 0; i < len(service.Backends); i++ {
			backendName := service.Backends[i]
			backend, err := e.ctx.GetBackend(serviceName, backendName)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("error getting backend %s from service %s", backendName, serviceName))
			}

			serviceBackendUptimeTotal.WithLabelValues(serviceName, backendName, backend.Options.IP.String(),
				fmt.Sprintf("%d", backend.Options.Port)).
				Set(backend.Metrics.Uptime.Seconds())

			serviceBackendHealth.WithLabelValues(serviceName, backendName, backend.Options.IP.String(),
				fmt.Sprintf("%d", backend.Options.Port)).
				Set(backend.Metrics.Health)

			serviceBackendStatus.WithLabelValues(serviceName, backendName, backend.Options.IP.String(),
				fmt.Sprintf("%d", backend.Options.Port)).
				Set(float64(backend.Metrics.Status))

			serviceBackendWeight.WithLabelValues(serviceName, backendName, backend.Options.IP.String(),
				fmt.Sprintf("%d", backend.Options.Port)).
				Set(float64(backend.Options.Weight))
		}
	}
	return nil
}
func RegisterPrometheusExporter(ctx *Context) {
	prometheus.MustRegister(NewExporter(ctx))
}
