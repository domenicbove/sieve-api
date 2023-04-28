.PHONY: build-api
build-api:
	eval $(minikube docker-env)
	docker build -t sieve-api:latest -f build/Dockerfile .

.PHONY: build-model
build-model:
	eval $(minikube docker-env)
	docker build -t model:latest -f model/Dockerfile .

.PHONY: deploy
deploy:
	kubectl apply -f build/deployment.yaml

.PHONY: cleanup
cleanup:
	kubectl delete -f build/deployment.yaml