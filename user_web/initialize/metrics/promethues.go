package metrics

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
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
