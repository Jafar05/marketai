package probes

import (
	"github.com/prometheus/client_golang/prometheus"
)

type ProbeCfg struct {
	ReadinessPath        string `mapstructure:"readinessPath" validate:"required,gte=0,lte=255"`
	LivenessPath         string `mapstructure:"livenessPath" validate:"required,gte=0,lte=255"`
	Port                 string `mapstructure:"port" validate:"required,gte=0,lte=255"`
	Pprof                string `mapstructure:"pprof"`
	PrometheusPath       string `mapstructure:"prometheusPath" validate:"required,gte=0,lte=255"`
	CheckIntervalSeconds int    `mapstructure:"checkIntervalSeconds" validate:"required,gte=0"`
}

type ProbeMetrics struct {
	LivenessSuccess  prometheus.Counter
	ReadinessSuccess prometheus.Counter
	ReadinessError   prometheus.Counter
}
