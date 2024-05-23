IMAGE_NAME := dcristobal/port-scan-exporter-app
CHART_NAME := port-scan-exporter
NAMESPACE := monitoring
CGO_ENABLED := 0 
GOOS := linux


.PHONY: build
build:
	go build -o /port-scan-exporter ./cmd

.PHONY: build-docker
build-docker:
	docker build -t $(IMAGE_NAME):latest .

.PHONY: push-docker
push-docker:
	docker push $(IMAGE_NAME):latest

.PHONY: deploy-chart
deploy-chart:
	helm install $(CHART_NAME) ./$(CHART_NAME) --namespace $(NAMESPACE)

.PHONY: upgrade-chart
upgrade-chart:
	helm upgrade $(CHART_NAME) ./$(CHART_NAME) --namespace $(NAMESPACE)

.PHONY: delete-chart
delete-chart:
	helm uninstall $(CHART_NAME) --namespace $(NAMESPACE)
