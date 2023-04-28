# Sieve Takehome
This is a simple example of an API to orchestrate a dummy ML Model on Kubernetes. The project is deployed on Minikube, but can be deployed on any K8s cluster!

### Prerequisites
- minikube
- kubectl cli

## Steps
1. Start Minikube
```
minikube start
```

2. Build API Image and Model Image (accessible within minikube)
```
make build-api
make build-model
```

3. Deploy Application Stack
```
make deploy
```

4. Port-foward the API
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
{"input":"whatever","latency":"117"}
```

## Test
While the ports are forward to trigger the test file run. (I'm using python3 cli)
```
python3 -m pip install requests
make test
```