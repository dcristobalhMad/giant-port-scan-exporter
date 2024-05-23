# Port Scan Exporter Documentation

## Overview

The Port Scan Exporter is a Kubernetes application written in Go that periodically scans the pods in a Kubernetes cluster to detect open ports. It exposes Prometheus metrics to monitor the state of open ports and the scanning process.

## Functionality

- **Port Scanning**: The exporter scans each pod in the cluster and checks for open ports.
- **Prometheus Metrics**: Exposes Prometheus metrics to monitor the number of open ports, scanned pods, scan duration, and errors encountered during scanning.
- **Configuration**: Supports configuration via environment variables for parameters such as port scan timeout, maximum port number, rescan interval, concurrency settings, and more.
- **Health Endpoint**: Provides a `/healthz` endpoint for basic health checks.
- **Logging**: Logs metrics and errors for monitoring and troubleshooting.

## Limitations and Possible Issues

1. **Scalability**: May not scale well in very large clusters.
2. **Concurrency and Resource Usage**: High concurrency can lead to high CPU and memory usage.
3. **Error Handling**: Basic error handling may make troubleshooting difficult.
4. **Network Policies**: Issues may arise with restrictive network policies.
5. **Security**: May raise security concerns due to intrusive port scanning.
6. **Configuration Management**: Limited flexibility in configuration management.
7. **Port Range**: Scans ports up to a configurable maximum, which may be excessive.

## Future Improvements

1. **Improved Scalability**: Optimize scanning algorithms for larger clusters.
2. **Enhanced Error Handling**: Implement detailed error logging and retry mechanisms.
3. **Advanced Configuration Management**: Use dynamic configuration systems.
4. **Dynamic Pod Discovery**: Implement dynamic pod discovery mechanisms.
5. **Metrics Enhancement**: Add more detailed metrics and historical tracking.
6. **Alerting and Monitoring**: Integrate with alerting systems and provide dashboards.
7. **User Interface**: Develop a web interface or CLI tool for interaction.
8. **Documentation and Testing**: Improve documentation and testing coverage.

## Usage

### Deployment manually

This Makefile is designed to facilitate the building, Dockerizing, and deploying of a Go-based port scan exporter application using Helm.

#### Variables

- **`IMAGE_NAME`**: The name of the Docker image to be created and pushed. In this case, it's set to `dcristobal/port-scan-exporter-app`.
- **`CHART_NAME`**: The name of the Helm chart used for deployment. Here, it's set to `port-scan-exporter`.
- **`NAMESPACE`**: The Kubernetes namespace where the Helm chart will be deployed. It's set to `monitoring`.
- **`CGO_ENABLED`**: A Go environment variable to disable CGO (set to 0).
- **`GOOS`**: The target operating system for the Go build (set to `linux`).

#### Targets

**`build`**

- **Purpose**: Compiles the Go application.
- **Command**: `go build -o /port-scan-exporter ./cmd`

**`build-docker`**

- **Purpose**: Builds the Docker image for the application.
- **Command**: `docker build -t $(IMAGE_NAME):latest .`

**`push-docker`**

- **Purpose**: Pushes the Docker image to a Docker registry.
- **Command**: `docker push $(IMAGE_NAME):latest`
  - `docker push`: The Docker command to push an image to a registry.
  - `$(IMAGE_NAME):latest`: The name and tag of the image to push.

**`deploy-chart`**

- **Purpose**: Deploys the Helm chart to the Kubernetes cluster.
- **Command**: `helm install $(CHART_NAME) ./$(CHART_NAME) --namespace $(NAMESPACE)`

**`upgrade-chart`**

- **Purpose**: Upgrades the Helm chart deployment in the Kubernetes cluster.
- **Command**: `helm upgrade $(CHART_NAME) ./$(CHART_NAME) --namespace $(NAMESPACE)`

**`delete-chart`**

- **Purpose**: Deletes the Helm chart deployment from the Kubernetes cluster.
- **Command**: `helm uninstall $(CHART_NAME) --namespace $(NAMESPACE)`

### Configuration

Configure the exporter using environment variables:

- `PORTSCAN_TIMEOUT_MS`: Port scan timeout in milliseconds (default: 150).
- `MAX_PORT`: Maximum port number to scan (default: 65535).
- `RESCAN_INTERVAL_MINUTES`: Interval between port scan cycles in minutes (default: 10).
- `PORTSCAN_WORKERS`: Number of workers for port scanning (default: 6).
- `MAX_PARALLEL_POD_SCANS`: Maximum number of parallel pod scans (default: 5).

### Monitoring

1. Ensure Prometheus is configured to scrape the `/metrics` endpoint.
2. Use Grafana or Prometheus to visualize and monitor the exposed metrics.
