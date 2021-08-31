import json
import random
import time
import math

import gevent
import locust
from websocket import create_connection


class ChannelTaskSet(locust.TaskSet):
    wait_time = locust.between(5, 10)

    def __init__(self, parent):
        super().__init__(parent)
        self.ws = create_connection('ws://localhost:22222/channel')

    def on_start(self):
        def _receive():
            while True:
                res = self.ws.recv()
                data = json.loads(res)
                res_time = time.time() * 1000

                locust.events.request_success.fire(
                    request_type='recv',
                    name='body',
                    response_time=res_time - data['regTime'],
                    response_length=len(res),
                )

        gevent.spawn(_receive)

    def on_quit(self):
        self.ws.close()

    @locust.task
    def send(self):
        if self.ws:
            data = {
                "regTime": round(time.time() * 1000),
                "body": "test",
            }
            body = json.dumps(data)
            self.ws.send(body)

            locust.events.request_success.fire(
                    request_type='send',
                    name='success',
                    response_time=0,
                    response_length=0,
                )


class ChatLocust(locust.HttpUser):
    tasks = [ChannelTaskSet]
