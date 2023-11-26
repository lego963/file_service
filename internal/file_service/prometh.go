package file_service

import "github.com/prometheus/client_golang/prometheus"

var (
	filesCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "imaginarium_files_total",
			Help: "Total number of files on disk.",
		},
		[]string{},
	)

	requestsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "imaginarium_requests_total",
			Help: "Total number of gRPC requests.",
		},
		[]string{"method"},
	)

	activeRequestsGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "imaginarium_active_requests",
			Help: "Number of active gRPC requests.",
		},
	)
)

func init() {
	prometheus.MustRegister(filesCounter)
	prometheus.MustRegister(requestsCounter)
	prometheus.MustRegister(activeRequestsGauge)
}
