import json
import time
import websockets
from cle import Cle
from config import Config
from tag import Tag
import numpy as np
import asyncio


def process_config(msg, CLE, cfg):
    cle = get_current_cle(msg, CLE)
    if cle:
        CLE.remove(cle)
    cle = Cle(msg, cfg)
    print("New CLE")    
    CLE.append(cle)

async  def process_TX(msg, cle, ws):
    for anchor in cle.anchors:
        if msg["data"]["receiver"] == anchor.master_ID:
            anchor.add_tx(msg)
            if anchor.data2sendflag:
                data2send = get_anchor_info(anchor)
                print('Process_TX data2send ', data2send)
                print("")
                await ws.send(json.dumps({'action':'SendToUserAnchorInfo','data':data2send,'apikey':apikey, 'organization': cle.organization}))
                anchor.data2sendflag = 0

async  def process_RX(msg, cle, ws):
        for anchor in cle.anchors:
            if msg["data"]["receiver"] == anchor.ID and msg["data"]["sender"] == anchor.master_ID:
                anchor.add_rx(msg)
                if anchor.data2sendflag:
                    data2send = get_anchor_info(anchor)
                    print('Process_RX data2send ', data2send)
                    print("")
                    await ws.send(json.dumps({'action':'SendToUserAnchorInfo','data':data2send,'apikey':apikey, 'organization': cle.organization}))
                    anchor.data2sendflag = 0
                break

async  def process_BLINK(msg, cle, ws):

    for anchor in cle.anchors:
        if msg["data"]["receiver"] == anchor.ID:
            break

    if cle.use_ref_tag and cle.ref_tag:
        ID = msg["data"]["sender"]
        if ID == cle.ref_tag["name"]:
            for anchor in cle.anchors:
                if msg["data"]["receiver"] == anchor.master_ID:
                    anchor.add_master_ref_tag(msg)
                    if anchor.data2sendflag:
                        data2send = get_anchor_info(anchor)
                        await ws.send(json.dumps({'action':'SendToUserAnchorInfo','data':data2send,'apikey':apikey, 'organization': cle.organization}))
                        anchor.data2sendflag = 0
            for anchor in cle.anchors:
                if msg["data"]["receiver"] == anchor.ID:
                    anchor.add_slave_ref_tag(msg)
                    if anchor.data2sendflag:
                        data2send = get_anchor_info(anchor)
                        await ws.send(json.dumps({'action':'SendToUserAnchorInfo','data':data2send,'apikey':apikey, 'organization': cle.organization}))
                        anchor.data2sendflag = 0
                    break
            return

    if anchor.sync_flag:
        msg["data"]["anchor_number"] = anchor.number
        msg["data"]["corrected_timestamp"] = anchor.correct_timestamp(msg["data"]["timestamp"])
        if cle.use_ref_tag and cle.ref_tag:
            msg["data"]["corrected_timestamp"] += anchor.ref_tag_correction
        match_flag = 0
        for tag in cle.tags:
            if msg["data"]["sender"] == tag.ID:
                match_flag = 1
                break

        if match_flag == 0:
            tag = Tag(msg, cle)
            print(f"New tag {tag.ID}")
            cle.tags.append(tag) 
        tag.add_data(msg)
        if tag.data2sendflag:
            data2send = get_tag_info(tag, cle)
            print('Process_BLINK data2send ', data2send)
            print("")
            await ws.send(json.dumps({'action':'SendToUser','data':data2send,'apikey':apikey, 'organization': cle.organization}))
            tag.data2sendflag = 0

def get_current_cle(msg, CLE):
    cle = []
    for cle in CLE:
        if cle.organization == msg["data"]["organization"] and cle.roomid == msg["data"]["roomid"]:
            print("GET CURRENT CLE")
            break
    return cle

def get_tag_info(tag, cle):
    tag_info = {}
    tag_info["type"] = "tag"
    tag_info["ID"] = tag.ID



    if tag.x_f is None and tag.y_f is None:
        tag.x_f = tag.x
        tag.y_f = tag.y
    else:
        tag.x_f = tag.x_f * tag.cle.cfg.alpha + tag.x * (1 - tag.cle.cfg.alpha)
        tag.y_f = tag.y_f * tag.cle.cfg.alpha + tag.y * (1 - tag.cle.cfg.alpha)
    tag_info["x_f"] = tag.x_f
    tag_info["y_f"] = tag.y_f

    tag_info["x"] = tag.x_f
    tag_info["y"] = tag.y_f
    tag_info["z"] = tag.h

    tag_info["dop"] = tag.DOP
    tag_info["lifetime"] = tag.lifetime
    tag_info["time"] = time.strftime("%d.%m.%Y %H:%M:%S", time.localtime(tag.lasttime))
    tag_info["anchors"] = tag.anchors_number_to_solve
    tag_info["static"] = tag.static
    tag_info["SOS"] = tag.SOS
    tag_info["organization"] = tag.organization
    tag_info["roomid"] = tag.roomid
    return tag_info

def get_anchor_info(anchor):
    anchor_info = {}
    anchor_info["type"] = "anchor"
    anchor_info["number"] = anchor.number
    anchor_info["ID"] = anchor.ID
    anchor_info["master_ID"] = anchor.master_ID
    anchor_info["role"] = anchor.Role
    anchor_info["x"] = anchor.x
    anchor_info["y"] = anchor.y
    anchor_info["z"] = anchor.z
    anchor_info["delta"] = float(anchor.X[0])
    anchor_info["delta_t"] = float(anchor.X[1])
    anchor_info["sqrt(Dx[1,1])"] = np.sqrt(float(anchor.Dx[0][0]))
    anchor_info["sqrt(Dx[2,2])"] = np.sqrt(float(anchor.Dx[1][1]))
    anchor_info["current_tx"] = float(anchor.current_tx)
    anchor_info["current_rx"] = float(anchor.current_rx)
    anchor_info["k_skip"] = int(anchor.k_skip)
    anchor_info["sync_flag"] = anchor.sync_flag
    anchor_info["ref_tag_correction"] = anchor.ref_tag_correction
    anchor_info["organization"] = anchor.organization
    anchor_info["roomid"] = anchor.roomid
    return anchor_info

async def math_client_handler():
    url = f"ws://{server_ip}:{server_port}"
    async with websockets.connect(url, ping_interval=None) as ws:
        await ws.send(json.dumps({"action": "Login", "login": "mathLogin", "password": "%wPp7VO6k7ump{BP4mu2rm4w?p|J5N%P", "roomid": "1"}))
        await asyncio.gather(math_client_sender(ws), math_client_receiver(ws))

async def math_client_sender(ws):
    while True:
        await asyncio.sleep(60)

async def math_client_receiver(ws):
    while True:
        message = json.loads(await ws.recv())
        print("MESSAGE FROM SERVER: " + str(message))
        if 'action' in message:
            if message["action"] == "RoomConfig":
                process_config(message, CLE, cfg)
            elif message["action"] == "CS_TX":
                cle = get_current_cle(message, CLE)
                if cle:
                    await process_TX(message, cle, ws)
            elif message["action"] == "CS_RX":
                cle = get_current_cle(message, CLE)
                if cle:
                    await process_RX(message, cle, ws)
            elif message["action"] == "BLINK":
                cle = get_current_cle(message, CLE)
                if cle:
                    await process_BLINK(message, cle, ws)
            elif message["action"] == "Login":
                global apikey
                if 'data' in message:
                    if 'apikey' in message["data"]:
                        apikey = message["data"]["apikey"]
                print("Login succsess")


if __name__ == "__main__":
    cfg = Config()
    CLE = []
    apikey = ""

    server_ip = "127.0.0.1"
    server_port = "8000"

    # try:
    #     with open('../NODE/NodeConfig.json') as file:
    #         node_config = json.load(file)
    #         server_ip = node_config["node_server_ip"]
    #         server_port = node_config["node_server_port"]
    # except:
    #     pass

    loop = asyncio.get_event_loop()
    loop.run_until_complete(math_client_handler())
