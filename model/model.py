import uuid
import time
import sys
import redis
import json

class Model:
    def __init__(self, job_id: str):

        # get job details and update status
        redis_client = redis.Redis(host='redis') # svc name in k8s
        job_details_bytes = redis_client.get(job_id)

        job_details = json.loads(job_details_bytes)
        job_details['status'] = "processing"

        job_details_bytes = json.dumps(job_details).encode('utf-8')

        redis_client.set(job_id, job_details_bytes)

        time.sleep(3)
        # time.sleep(35)
        self.job_id = job_id
        self.redis_client = redis_client
        self.return_val = "world" + str(uuid.uuid4())

    def predict(self, hello: str):
        time.sleep(3)
        # time.sleep(10)

        # get job details from redis and update output and status
        job_details_bytes = self.redis_client.get(self.job_id)

        job_details = json.loads(job_details_bytes)
        job_details['status'] = "done"
        job_details['output'] = self.return_val

        job_details_bytes = json.dumps(job_details).encode('utf-8')
        self.redis_client.set(self.job_id, job_details_bytes)

        return {"output": self.return_val, "input": hello}

model1 = Model(job_id=sys.argv[1])
out = model1.predict(hello="Prince") # unsure what the hello input is meant to be
