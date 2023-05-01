.PHONY: build-api
build-api:
	docker build -t sieve-api:latest -f build/Dockerfile .

.PHONY: build-model
build-model:
	docker build -t model:latest -f model/Dockerfile .

.PHONY: deploy
deploy:
	kubectl apply -f build/deployment.yaml

.PHONY: cleanup
cleanup:
	kubectl delete -f build/deployment.yaml

.PHONY: test
test:
	python3 test/test.py