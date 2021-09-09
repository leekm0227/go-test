import json
import random
import time
import math

import gevent
import locust
from websocket import create_connection


class ChannelTaskSet(locust.TaskSet):
    wait_time = locust.between(25, 30)

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

                try:
                    if data["payloadType"] == "INIT":
                        self.id = data["id"]
                        self.users = data["players"]
                    elif data["payloadType"] == "MOVE":
                        self.users[data["player"]["id"]]["pos"] = data["player"]["pos"]
                    elif data["payloadType"] == "SPAWN":
                        self.users[data["player"]["id"]] = data["player"]
                    elif data["payloadType"] == "ATTACK":
                        self.users[data["player"]["id"]]["hp"] = data["player"]["hp"]
                    elif data["payloadType"] == "DEAD":
                        del self.users[data["sessionId"]]

                    locust.events.request_success.fire(
                        request_type='recv',
                        name='pos',
                        response_time=res_time - data['regTime'],
                        response_length=len(res),
                    )
                except Exception as e:
                    locust.events.request_failure.fire(
                        request_type='recv',
                        name='fail',
                        response_time=0,
                        response_length=0,
                        exception=(data, e)
                    )

        gevent.spawn(_receive)

    def on_quit(self):
        self.ws.close()

    @locust.task
    def send(self):
        if self.id != "":
            try:
                # attack or move
                if random.choice([True, False]):
                    self.target = ""

                    for k in self.users.keys():
                        if self.id != k:
                            pos1 = self.users[self.id]["pos"]
                            pos2 = self.users[k]["pos"]
                            x = pos1[0] - pos2[0]
                            y = pos1[1] - pos2[1]
                            dis = math.sqrt(abs(x*x) + abs(y*y))

                            if dis < 2:
                                self.target = k
                                break

                    if self.target != "":
                        data = {
                            "payloadType": 3,
                            "receiveType": 1,
                            "regTime": round(time.time() * 1000),
                            "targetId": self.target,
                        }
                        body = json.dumps(data)
                        # self.ws.send(body)
                        locust.events.request_success.fire(
                            request_type='send',
                            name='success',
                            response_time=0,
                            response_length=0,
                        )
                        return

                # set dir
                if random.choice([True, False]):
                    if random.choice([True, False]):
                        direction = [-1, 0]
                    else:
                        direction = [1, 0]
                else:
                    if random.choice([True, False]):
                        direction = [0, -1]
                    else:
                        direction = [0, 1]

                data = {
                    "payloadType": 2,
                    "receiveType": 1,
                    "regTime": round(time.time() * 1000),
                    "dir": direction,
                }
                body = json.dumps(data)
                # self.ws.send(body)
                locust.events.request_success.fire(
                    request_type='send',
                    name='success',
                    response_time=0,
                    response_length=0,
                )
            except Exception as e:
                locust.events.request_failure.fire(
                    request_type='send',
                    name='fail',
                    response_time=0,
                    response_length=0,
                    exception=e
                )
        else:
            data = {"payloadType": 1, "regTime": round(time.time() * 1000)}
            body = json.dumps(data)
            # self.ws.send(body)


class ChatLocust(locust.HttpUser):
    tasks = [ChannelTaskSet]
