package main

import (
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"giant-port-scan-exporter/m/internal/pkg"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = clientcmd.RecommendedHomeFile
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	return kubernetes.NewForConfig(config)
}

func main() {
	pkg.PortscanTimeout = pkg.GetEnvDuration("PORTSCAN_TIMEOUT_MS", 150*time.Millisecond)
	pkg.MaxPort = pkg.GetEnvInt("MAX_PORT", 65535)
	pkg.RescanInterval = pkg.GetEnvDuration("RESCAN_INTERVAL_MINUTES", 10*time.Minute)
	pkg.PortscanWorkers = pkg.GetEnvInt("PORTSCAN_WORKERS", 6)
	pkg.MaxParallelPodScans = pkg.GetEnvInt("MAX_PARALLEL_POD_SCANS", 5)

	clientset, err := getClientset()
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	go func() {
		for {
			var wg sync.WaitGroup
			semaphore := make(chan struct{}, pkg.PortscanWorkers)
			wg.Add(1)
			go pkg.ScanPodPorts(clientset, &wg, semaphore)
			wg.Wait()
			time.Sleep(pkg.RescanInterval)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Println("Starting exporter on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
