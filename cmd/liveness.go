package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"time"

	connlib "github.com/kubernetes-csi/csi-lib-utils/connection"
	"github.com/kubernetes-csi/csi-lib-utils/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	csiAddress = flag.String("csi-address", "/run/csi/socket", "Address of the CSI driver socket.")
	port       = flag.String("port", "8080", "TCP port for listening requests")
	endpoint   = flag.String("endpoint", "/metrics", "endpoint for requests")
	pollTime   = flag.Int("poll-time", 60, "seconds for between liveness polls")

	liveness = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "csi",
		Name:      "liveness",
		Help:      "Liveness Probe",
	})
)

func getLiveness() {

	csiConn, err := connlib.Connect(*csiAddress)
	if err != nil {
		// connlib should retry forever so a returned error should mean
		// the grpc client is misconfigured rather than an error on the network
		log.Fatalf("failed to establish connection to CSI driver: %v", err)
	}

	log.Printf("calling CSI driver to discover driver name")
	csiDriverName, err := rpc.GetDriverName(context.Background(), csiConn)
	if err != nil {
		log.Fatalf("failed to get CSI driver name: %v", err)
	}
	log.Printf("CSI driver name: %q", csiDriverName)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	log.Printf("Sending probe request to CSI driver %q", csiDriverName)
	ready, err := rpc.Probe(ctx, csiConn)
	if err != nil {
		liveness.Set(0)
		log.Printf("health check failed: %v", err)
		return
	}

	if !ready {
		liveness.Set(0)
		log.Printf("driver responded but is not ready")
		return
	}
	liveness.Set(1)
	log.Printf("Health check succeeded")
}

func recordLiveness() {
	//register promethues metrics
	prometheus.MustRegister(liveness)

	//get liveness periodically
	go func() {
		for {
			getLiveness()
			// wait to poll metrics again
			time.Sleep(time.Duration(*pollTime) * time.Second)
		}
	}()
}

func main() {
	flag.Parse()
	// start liveness collection
	recordLiveness()
	// start up prometheus endpoint
	addr := net.JoinHostPort("0.0.0.0", *port)
	http.Handle(*endpoint, promhttp.Handler())
	http.ListenAndServe(addr, nil)
}
