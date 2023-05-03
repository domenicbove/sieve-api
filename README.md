# Sieve Takehome
This is a simple example of an API to orchestrate a mock ML Model on Kubernetes and store details on Redis. The project is deployed on Minikube, but can be deployed on any K8s cluster!

Note: this project will run in the default namespace

### Prerequisites
- docker
- minikube
- kubectl cli

## Steps
1. Start Minikube
```
minikube start
```

2. Make new Docker image Accessible within Minikube
```
eval $(minikube docker-env)
```

3. Build API and Model Images
```
make build-api
make build-model
```

3. Deploy Application Stack
```
make deploy
```

4. Port-foward the API in a new terminal
```
kubectl port-forward $(kubectl get pod -l app=sieve-api  -o=jsonpath='{.items[0].metadata.name}') 8080:8080
```

5. Trigger a new Model
```
# curl -X POST http://localhost:8080/push -H 'Content-Type: application/json' -d '{"input": "dommy"}'
{"id":"edbf2f12-9148-4623-8f6e-ec35d254dbf4"}
```

6. Query Status
```
# curl http://localhost:8080/status/<id>
{"status":"queued"}
```

7. Query Result
```
# curl http://localhost:8080/data/<id>
{"input":"dommy","latency":"48","output":"world0d2bfd51-347a-4221-9b6a-60a9e7873ee5"}
```

## Test
While the API port is forwarded, to trigger the test file run. (I'm using python3 cli)
```
python3 -m pip install requests
make test
```

After a test run, you can clear out the Completed Model Pods with:
```
kubectl delete jobs `kubectl get jobs -o custom-columns=:.metadata.name`
```

## Teardown
Finally, to remove the application stack
```
make cleanup
```