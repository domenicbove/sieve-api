.PHONY: build-minikube
build-minikube:
	eval $(minikube docker-env)
	docker build -t sieve-api:latest -f build/Dockerfile .

.PHONY: build
build: 
	docker build -t sieve-api:latest -f build/Dockerfile .

.PHONY: deploy
deploy:
	kubectl apply -f build/deployment.yaml