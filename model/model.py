import uuid
import time
import sys
import redis
import json

class Model:
    def __init__(self):
        time.sleep(3)
        # time.sleep(35)
        self.return_val = "world" + str(uuid.uuid4())

    def predict(self, hello: str):
        time.sleep(10)
        return {"output": self.return_val, "input": hello}

job_id = sys.argv[1]
input = sys.argv[2]

model1 = Model()

# after creating the model, update job status in redis
redis_client = redis.Redis(host='redis') # svc name in k8s
job_details_bytes = redis_client.get(job_id)

job_details = json.loads(job_details_bytes)
job_details['status'] = "processing"

job_details_bytes = json.dumps(job_details).encode('utf-8')
redis_client.set(job_id, job_details_bytes)

out = model1.predict(hello=input)

# after model completes, update job details in redis
job_details['status'] = "finished"
job_details['output'] = out["output"]
job_details['endTime'] = int(time.time())

job_details_bytes = json.dumps(job_details).encode('utf-8')
redis_client.set(job_id, job_details_bytes)
