import uuid
import time
import sys
import redis
import json
import os

from kubernetes import client, config

REDIS_JOB_QUEUE_NAME = "jobs"
REDIS_HOSTNAME = "redis"
JOB_STATUS_FINISHED = "finished"
JOB_STATUS_PROCESSING = "processing"

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
    job_details['status'] = JOB_STATUS_PROCESSING

    job_details_bytes = json.dumps(job_details).encode('utf-8')
    redis_client.set(job_id, job_details_bytes)

    # execute model with input
    out = model1.predict(hello=job_details['input'])

    # update job details in redis
    job_details['status'] = JOB_STATUS_FINISHED
    job_details['output'] = out["output"]
    job_details['endTime'] = int(time.time())

    job_details_bytes = json.dumps(job_details).encode('utf-8')
    redis_client.set(job_id, job_details_bytes)


redis_client = redis.Redis(host=REDIS_HOSTNAME)

# if there are no jobs in queue, exit
if redis_client.llen(REDIS_JOB_QUEUE_NAME) == 0:
    sys.exit()

print("Setting up Model")
model1 = Model()

config.load_incluster_config()
v1 = client.CoreV1Api()

# hostname env var matches podname in k8s pods
patch_response = v1.patch_namespaced_pod(os.getenv('HOSTNAME'), "default", body={
  "metadata":{"labels":{"setup-complete": "true"}}
})
print("Pod annotation added. status='%s'" % str(patch_response.status))

while True:
    job = redis_client.lpop(REDIS_JOB_QUEUE_NAME)
    if job is None:
        sys.exit()
    job_id = job.decode('utf-8')
    runJob(redis_client, job_id)
