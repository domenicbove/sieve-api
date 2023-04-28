import uuid
import time
import redis
from redis.commands.json.path import Path

class Model:
    def __init__(self, job_id: str):
        redis_client = redis.Redis()

        job_details = redis_client.json().get(job_id)
        print(job_details)

        job_details['status'] = "processing"
        redis_client.json().set(job_id, Path.root_path(), job_details)

        time.sleep(1)
        # time.sleep(35)
        self.job_id = job_id
        self.redis_client = redis_client
        self.return_val = "world" + str(uuid.uuid4())

    def predict(self, hello: str):
        time.sleep(1)
        # time.sleep(10)

        job_details = self.redis_client.json().get(self.job_id)
        job_details['status'] = "done"
        job_details['output'] = self.return_val
        self.redis_client.json().set(self.job_id, Path.root_path(), job_details)

        return {"output": self.return_val, "input": hello}

model1 = Model(job_id="xxx")
out = model1.predict(hello="Prince")
print(out)
