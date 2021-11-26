.PHONY: build cover deploy start test test-integration

image = `aws lightsail get-container-images --service-name canvas | jq -r '.containerImages[].image' | grep app | head -n 1`
prometheus_image = `aws lightsail get-container-images --service-name canvas | jq -r '.containerImages[].image' | grep prometheus | head -n 1`

build:
	docker build -t canvas .
	docker build -t canvas_prometheus prometheus

cover:
	go tool cover -html=cover.out

deploy:
	aws lightsail push-container-image --service-name canvas --label app --image canvas
	aws lightsail push-container-image --service-name canvas --label prometheus --image canvas_prometheus
	jq <containers.json ".app.image=\"$(image)\"" >containers2.json
	mv containers2.json containers.json
	jq <containers.json ".prometheus.image=\"$(prometheus_image)\"" >containers2.json
	mv containers2.json containers.json
	aws lightsail create-container-service-deployment --service-name canvas \
		--containers file://containers.json \
		--public-endpoint '{"containerName":"app","containerPort":8080,"healthCheck":{"path":"/health"}}'

start:
	go run cmd/server/*.go

test:
	go test -coverprofile=cover.out -short ./...

test-integration:
	go test -coverprofile=cover.out -p 1 ./...
