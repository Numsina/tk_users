package metrics

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

func InitPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":9988", nil)
		if err != nil {
			log.Println("prometheus server error:", err)
			return
		}
	}()
}
