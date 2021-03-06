package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Exporting in port 9338 because it is one of the
	// free exporter ports. For more info visit:
	// https://github.com/prometheus/prometheus/wiki/Default-port-allocations
	http.Handle("/metrics", promhttp.Handler())
	log.Info("Begining to serve on port 9392")
	go func() {
		self_update()
	}()
	log.Fatal(http.ListenAndServe(":9392", nil))
}
