import uuid
import time
import sys
import redis
import json

class Model:
    def __init__(self):
        time.sleep(35)
        self.return_val = "world" + str(uuid.uuid4())

    def predict(self, hello: str):
        time.sleep(10)
        return {"output": self.return_val, "input": hello}

def runJob(redis_client, job_id):
    print("Executing Model for Job: " + job_id)

    # after creating the model, update job status in redis
    job_details_bytes = redis_client.get(job_id)

    job_details = json.loads(job_details_bytes)
    job_details['status'] = "processing"

    job_details_bytes = json.dumps(job_details).encode('utf-8')
    redis_client.set(job_id, job_details_bytes)

    # execute model with input
    out = model1.predict(hello=job_details['input'])

    # update job details in redis
    job_details['status'] = "finished"
    job_details['output'] = out["output"]
    job_details['endTime'] = int(time.time())

    job_details_bytes = json.dumps(job_details).encode('utf-8')
    redis_client.set(job_id, job_details_bytes)


redis_client = redis.Redis(host='redis')

# if there are no jobs in queue, exit
if redis_client.llen('jobs') == 0:
    sys.exit()

print("Setting up Model")
model1 = Model()

while True:
    job = redis_client.lpop('jobs')
    if job is None:
        sys.exit()
    job_id = job.decode('utf-8')
    runJob(redis_client, job_id)
