import json
import random
import time
import math

import gevent
import locust
from websocket import create_connection


class ChannelTaskSet(locust.TaskSet):
    wait_time = locust.between(0.001, 0.001)

    def __init__(self, parent):
        super().__init__(parent)
        self.ws = create_connection('ws://localhost:22222/channel')
        self.id = ""
        self.target = ""
        self.users = {}

    def on_start(self):
        def _receive():
            while True:
                res = self.ws.recv()
                data = json.loads(res)
                res_time = time.time() * 1000

                locust.events.request_success.fire(
                        request_type='recv',
                        name='pos',
                        response_time=res_time - data['regTime'],
                        response_length=len(res),
                )

        gevent.spawn(_receive)

    def on_quit(self):
        self.ws.close()

    @locust.task
    def send(self):
        data = {"payloadType": 1, "regTime": round(time.time() * 1000)}
        body = json.dumps(data)
        self.ws.send(body)

class ChatLocust(locust.HttpUser):
    tasks = [ChannelTaskSet]
