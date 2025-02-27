package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Namespace  string
	Subsystem  string
	Name       string
	Help       string
	InstanceID string
}

func NewMetrics(namespace string, instanceID string, name string, subsystem string, help string) *Metrics {
	return &Metrics{Namespace: namespace, InstanceID: instanceID, Name: name, Subsystem: subsystem, Help: help}
}

func (m *Metrics) Build() gin.HandlerFunc {

	// 统计请求的methods, path, status， 请求的时间
	summery := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: m.Namespace,
		Subsystem: m.Subsystem,
		Name:      m.Name + "_resp_time",
		Help:      m.Help,
		ConstLabels: map[string]string{
			"instance_id": m.InstanceID,
		},
		Objectives: map[float64]float64{
			0.5:   0.1,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"method", "path", "status"})
	prometheus.Register(summery)

	// 统计活跃请求
	gague := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: m.Namespace,
		Subsystem: m.Subsystem,
		Name:      m.Name + "_active_request",
		Help:      m.Help,
		ConstLabels: map[string]string{
			"instance_id": m.InstanceID,
		},
	})
	prometheus.Register(gague)

	// 统计请求总数
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: m.Namespace,
		Subsystem: m.Subsystem,
		Name:      m.Name + "_all_request",
		Help:      m.Help,
		ConstLabels: map[string]string{
			"instance_id": m.InstanceID,
		},
	})
	prometheus.Register(counter)
	return func(c *gin.Context) {
		start := time.Now()
		gague.Inc()
		counter.Inc()
		defer func() {
			gague.Dec()
			method := c.Request.Method
			path := c.FullPath()

			if path == "" {
				path = "unknown"
			}
			end := time.Since(start).Milliseconds()
			status := c.Writer.Status()
			summery.WithLabelValues(method, path, strconv.Itoa(status)).Observe(float64(end))
		}()
		c.Next()

	}
}
