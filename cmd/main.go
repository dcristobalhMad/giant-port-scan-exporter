package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	openPorts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "open_ports",
			Help: "Number of open ports on a pod",
		},
		[]string{"namespace", "pod", "port"},
	)

	portscanTimeout     time.Duration
	maxPort             int
	rescanInterval      time.Duration
	portscanWorkers     int
	maxParallelPodScans int
)

func init() {
	prometheus.MustRegister(openPorts)
}

func getEnvInt(env string, fallback int) int {
	if value, ok := os.LookupEnv(env); ok {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

func getEnvDuration(env string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(env); ok {
		if intValue, err := strconv.Atoi(value); err == nil {
			return time.Duration(intValue) * time.Millisecond
		}
	}
	return fallback
}

func scanPort(ip string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func logMetric(namespace, pod, port string, value float64) {
	log.Printf("Metric - namespace: %s, pod: %s, port: %s, open: %f", namespace, pod, port, value)
}

func scanPodPorts(clientset *kubernetes.Clientset, wg *sync.WaitGroup, semaphore chan struct{}) {
	defer wg.Done()
	semaphore <- struct{}{}
	defer func() { <-semaphore }()

	ctx := context.Background()
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Error listing pods: %v", err)
		return
	}

	var podWg sync.WaitGroup
	podSemaphore := make(chan struct{}, maxParallelPodScans)

	for _, pod := range pods.Items {
		if pod.Spec.HostNetwork {
			continue
		}
		podWg.Add(1)
		go func(pod corev1.Pod) {
			defer podWg.Done()
			podSemaphore <- struct{}{}
			defer func() { <-podSemaphore }()

			for _, container := range pod.Spec.Containers {
				for _, port := range container.Ports {
					portNum := int(port.ContainerPort)
					if portNum > maxPort {
						continue
					}
					isOpen := 0.0
					if scanPort(pod.Status.PodIP, portNum, portscanTimeout) {
						isOpen = 1.0
					}
					openPorts.WithLabelValues(pod.Namespace, pod.Name, strconv.Itoa(portNum)).Set(isOpen)
					logMetric(pod.Namespace, pod.Name, strconv.Itoa(portNum), isOpen)
				}
			}
		}(pod)
	}
	podWg.Wait()
}

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
	portscanTimeout = getEnvDuration("PORTSCAN_TIMEOUT_MS", 150*time.Millisecond)
	maxPort = getEnvInt("MAX_PORT", 65535)
	rescanInterval = getEnvDuration("RESCAN_INTERVAL_MINUTES", 10*time.Minute)
	portscanWorkers = getEnvInt("PORTSCAN_WORKERS", 6)
	maxParallelPodScans = getEnvInt("MAX_PARALLEL_POD_SCANS", 5)

	clientset, err := getClientset()
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	go func() {
		for {
			var wg sync.WaitGroup
			semaphore := make(chan struct{}, portscanWorkers)
			wg.Add(1)
			go scanPodPorts(clientset, &wg, semaphore)
			wg.Wait()
			time.Sleep(rescanInterval)
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
