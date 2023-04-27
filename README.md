```
minikube start
```


To Build, should be accessible within minikube
```
# need to run this before building to make the image accessible by minikube i think
eval $(minikube docker-env)
docker build -t sieve-api .
```

To Deploy on minikube

To run redis # should put this into deployment
```
docker run --name redis-test-instance -p 6379:6379 -d redis

kubectl run redis --image=redis:latest --replicas=1 --port=6379
```

TODO add makefile

```
curl -X POST http://localhost:8080/push \\n   -H 'Content-Type: application/json' \\n   -d '{"input": "whatever" }'
curl http://localhost:8080/status/2381ac17-8c72-4513-895d-1a9f0d538fe0
curl http://localhost:8080/data/2381ac17-8c72-4513-895d-1a9f0d538fe0
kubectl port-forward sieve-api-69dfb57d48-vmbkw 8080:8080
```



