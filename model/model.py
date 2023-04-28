import uuid
import time
import redis

class Model:
    # TODO add jobId as parameter
    def __init__(self):
        time.sleep(1)
        # time.sleep(35)
        self.return_val = "world" + str(uuid.uuid4())

    def predict(self, hello: str):
        time.sleep(1)
        # time.sleep(10)
        redis_client = redis.Redis()
        redis_client.mset({"Croatia": "Zagreb", "Bahamas": "Nassau"})
        print(redis_client.get("Bahamas"))
        return {"output": self.return_val, "input": hello}

model1 = Model()
out = model1.predict(hello="Prince")
print(out)
