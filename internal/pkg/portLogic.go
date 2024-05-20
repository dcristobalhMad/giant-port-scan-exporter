package pkg

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	openPorts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "open_ports",
			Help: "Number of open ports on a pod",
		},
		[]string{"namespace", "pod", "port"},
	)
	RescanInterval      time.Duration
	PortscanWorkers     int
	MaxParallelPodScans int
	PortscanTimeout     time.Duration
	MaxPort             int
)

func Init() {
	prometheus.MustRegister(openPorts)
}

func ScanPort(ip string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func ScanPodPorts(clientset *kubernetes.Clientset, wg *sync.WaitGroup, semaphore chan struct{}) {
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
	podSemaphore := make(chan struct{}, MaxParallelPodScans)

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
					if portNum > MaxPort {
						continue
					}
					if ScanPort(pod.Status.PodIP, portNum, PortscanTimeout) {
						openPorts.WithLabelValues(pod.Namespace, pod.Name, strconv.Itoa(portNum)).Set(1)
					} else {
						openPorts.WithLabelValues(pod.Namespace, pod.Name, strconv.Itoa(portNum)).Set(0)
					}
				}
			}
		}(pod)
	}
	podWg.Wait()
}
