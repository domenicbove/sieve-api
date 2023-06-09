import requests
import time

API_URL = "http://localhost:8080"
API_PUSH = "http://localhost:8080" + "/push"
API_STATUS = "http://localhost:8080" + "/status"
API_DATA = "http://localhost:8080" + "/data"
ids = []
latencies = []
outputs = []
for i in range(5):
    r = requests.post(API_PUSH, params={"input": "test"})
    _id = r.json()["id"]
    ids.append(_id)

time.sleep(50)
for i in range(5):
    r = requests.post(API_PUSH, params={"input": "test"})
    _id = r.json()["id"]
    ids.append(_id)

for _id in ids:
    finished = False
    while not finished:
        r = requests.get(API_STATUS+f"/{str(_id)}")
        if r.json()["status"] == "finished":
            finished = True
        else:
            time.sleep(1)
    r = requests.get(API_DATA+f"/{str(_id)}")
    latencies.append(r.json()["latency"])
    outputs.append(r.json()["output"])

print(ids)
print(latencies)
print(outputs)