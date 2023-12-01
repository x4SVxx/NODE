package Anchor

import (
	"NODE/Logger"
	"NODE/ReportsAndMessages"
	"NODE/ServerForMath"
	"encoding/json"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

type anchor_struct struct {
	ip           string
	id           string
	number       float64
	masternumber float64
	role         string
	lag          float64
	adrx         float64
	adtx         float64
	x            float64
	y            float64
	z            float64
	connection   net.Conn
}

var anchors_array []anchor_struct

func CreateAnchor(data map[string]interface{}) {
	anchors_array = append(anchors_array, anchor_struct{
		ip:           data["ip"].(string),
		number:       data["number"].(float64),
		masternumber: data["masternumber"].(float64),
		role:         data["role"].(string),
		lag:          data["lag"].(float64),
		adrx:         data["adrx"].(float64),
		adtx:         data["adtx"].(float64),
		x:            data["x"].(float64),
		y:            data["y"].(float64),
		z:            data["z"].(float64),
	})
}

func ConnectAnchors(server_connection *websocket.Conn) {
	for i := 0; i < len(anchors_array); i++ {
		Connect(&anchors_array[i], server_connection)
	}

}

func SetRfConfigAnchors(rf_config map[string]interface{}, server_connection *websocket.Conn) {
	for i := 0; i < len(anchors_array); i++ {
		SetRfConfig(&anchors_array[i], rf_config, server_connection)
	}
}

func DisConnectAnchors(server_connection *websocket.Conn) {
	for i := 0; i < len(anchors_array); i++ {
		DisConnect(&anchors_array[i], server_connection)
	}
}

func ClearAnchors() {
	var anchors_array_nil []anchor_struct
	anchors_array = anchors_array_nil
}

func StartSpamAnchors(apikey string, name string, clientid string, roomid string, organization string, roomname string, independent_flag string, connect_math_flag string, start_spam_flag bool, rf_config map[string]interface{}, server_connection *websocket.Conn) {
	for i := 0; i < len(anchors_array); i++ {
		StartSpam(&anchors_array[i], server_connection)
		go Handler(apikey, name, clientid, roomid, organization, independent_flag, connect_math_flag, &anchors_array[i], server_connection)
	}
	go CheckAnchors(apikey, name, clientid, roomid, organization, independent_flag, connect_math_flag, start_spam_flag, rf_config, server_connection)
}

func StopSpamAnchors(server_connection *websocket.Conn) {
	for i := 0; i < len(anchors_array); i++ {
		StopSpam(&anchors_array[i], server_connection)
	}
}

func SetRoomConfigToMath(ref_tag_config map[string]interface{}, apikey string, name string, clientid string, roomid string, organization string, roomname string, independent_flag string, connect_math_flag string, server_connection *websocket.Conn) {
	var map_anchors []map[string]interface{}
	for i := 0; i < len(anchors_array); i++ {
		anchor := map[string]interface{}{
			"ip":           anchors_array[i].ip,
			"id":           anchors_array[i].id,
			"number":       anchors_array[i].number,
			"masternumber": anchors_array[i].masternumber,
			"role":         anchors_array[i].role,
			"lag":          anchors_array[i].lag,
			"adrx":         anchors_array[i].adrx,
			"adtx":         anchors_array[i].adtx,
			"x":            anchors_array[i].x,
			"y":            anchors_array[i].y,
			"z":            anchors_array[i].z,
		}
		map_anchors = append(map_anchors, anchor)
	}
	if connect_math_flag == "true" || independent_flag == "true" {
		ServerForMath.RoomAndReftagConfig(map_anchors, ref_tag_config)
	} else {
		MessageToServer(map[string]interface{}{"action": "RoomConfig", "apikey": apikey, "clientid": clientid, "organization": organization, "roomid": roomid, "roomname": roomname, "anchors": map_anchors, "ref_tag_config": ref_tag_config}, server_connection)
	}
}

func SendToMath(message map[string]interface{}, apikey string, name string, clientid string, roomid string, organization string, independent_flag string, connect_math_flag string, server_connection *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : SendToMath", err)
		}
	}()
	if message["type"] != "Unknow" {
		if independent_flag == "true" || connect_math_flag == "true" {
			math_map_message := map[string]interface{}{}
			if message["type"] == "CS_RX" || message["type"] == "CS_TX" {
				math_map_message["action"] = message["type"]
				math_map_message["data"] = map[string]interface{}{
					"apikey":       apikey,
					"orgname":      name,
					"organization": organization,
					"clientid":     clientid,
					"roomid":       roomid,
					"type":         message["type"],
					"timestamp":    message["timestamp"],
					"receiver":     message["receiver"],
					"sender":       message["sender"],
					"seq":          message["seq"],
				}
			}
			if message["type"] == "BLINK" {
				math_map_message["action"] = message["type"]
				math_map_message["data"] = map[string]interface{}{
					"apikey":       apikey,
					"orgname":      name,
					"organization": organization,
					"clientid":     clientid,
					"roomid":       roomid,
					"type":         message["type"],
					"timestamp":    message["timestamp"],
					"receiver":     message["receiver"],
					"sender":       message["sender"],
					"sn":           message["sn"],
					"state":        message["state"],
				}
			}
			json_math_message, _ := json.Marshal(math_map_message)
			Logger.Logger(string(json_math_message), nil)

			ServerForMath.MessageToMath(math_map_message)
		} else {
			if server_connection != nil {
				math_map_message := map[string]interface{}{
					"action":       "SendToMath",
					"apikey":       apikey,
					"orgname":      name,
					"organization": organization,
					"clientid":     clientid,
					"roomid":       roomid,
					"type":         message["type"],
					"timestamp":    message["timestamp"],
					"receiver":     message["receiver"],
					"sender":       message["sender"],
				}
				if message["type"] == "CS_RX" || message["type"] == "CS_TX" {
					math_map_message["seq"] = message["seq"]
				}
				if message["type"] == "BLINK" {
					math_map_message["sn"] = message["sn"]
				}
				json_math_message, _ := json.Marshal(math_map_message)
				Logger.Logger(string(json_math_message), nil)

				MessageToServer(math_map_message, server_connection)
			}
		}
	}
}

func Handler(apikey string, name string, clientid string, roomid string, organization string, independent_flag string, connect_math_flag string, anchor *anchor_struct, server_connection *websocket.Conn) {
	break_point := false
	for {
		if break_point {
			return
		}
		func() {
			defer func() {
				if err := recover(); err != nil {
					break_point = true
					Logger.Logger("ERROR : AnchorHandler", err)
					MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: AnchorHandler"}, server_connection)
					if anchor.connection != nil {
						anchor.connection.Close()
						anchor.connection = nil
					}
					anchor.id = ""
				}
			}()
			buffer_header := make([]byte, 3)
			anchor.connection.Read(buffer_header)
			number_of_bytes := buffer_header[1]
			buffer_anchor_message := make([]byte, number_of_bytes)
			anchor.connection.Read(buffer_anchor_message)
			buffer_ending := make([]byte, 3)
			anchor.connection.Read(buffer_ending)
			message := ReportsAndMessages.DecodeAnchorMessage(buffer_anchor_message)
			message["receiver"] = anchor.id
			if message["type"] == "CS_TX" {
				message["sender"] = message["receiver"]
			}
			SendToMath(message, apikey, name, clientid, roomid, organization, independent_flag, connect_math_flag, server_connection)
		}()
	}
}

func Connect(anchor *anchor_struct, server_connection *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : AnchorConnect", err)
			if server_connection != nil {
				MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: AnchorConnect"}, server_connection)
			}
		}
	}()
	anchor_connection, err := net.Dial("tcp", anchor.ip+":"+"3000")
	if err != nil {
		Logger.Logger("ERROR : AnchorConnect", err)
		if server_connection != nil {
			MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: AnchorConnect " + anchor.ip}, server_connection)
		}
	} else {
		buffer_skip := make([]byte, 3)
		anchor_connection.Read(buffer_skip)
		buffer_anchor_connect := make([]byte, 500)
		anchor_connection.Read(buffer_anchor_connect)
		anchor.connection = anchor_connection
		anchor.id = ReportsAndMessages.DecodeAnchorMessage(buffer_anchor_connect)["receiver"].(string)
		Logger.Logger("SUCCESS : AnchorConnect "+anchor.ip+" "+anchor.id, nil)
		if server_connection != nil {
			MessageToServer(map[string]interface{}{"action": "Success", "data": "Sucess: AnchorConnect " + anchor.ip}, server_connection)
		}

	}
}

func DisConnect(anchor *anchor_struct, server_connection *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : AnchorDisConnect", err)
			if server_connection != nil {
				MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: AnchorDisConnect"}, server_connection)
			}
		}
	}()
	if anchor.connection != nil {
		anchor.connection.Close()
		Logger.Logger("SUCCESS : AnchorDisConnect "+anchor.ip, nil)
		if server_connection != nil {
			MessageToServer(map[string]interface{}{"action": "Success", "data": "Sucess: AnchorDisConnect " + anchor.ip}, server_connection)
		}
	}
}

func SetRfConfig(anchor *anchor_struct, rf_config map[string]interface{}, server_connection *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : AnchorSetRfConfig", err)
			if server_connection != nil {
				MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: AnchorSetRfConfig"}, server_connection)
			}
		}
	}()
	var PRF map[int]int = map[int]int{
		16: 1,
		64: 2,
	}
	var DATARATE map[float64]int = map[float64]int{
		110: 0,
		850: 1,
		6.8: 2,
	}
	var PREAMBLE_LEN map[int]int = map[int]int{
		64:   int(0x04),
		128:  int(0x14),
		256:  int(0x24),
		512:  int(0x34),
		1024: int(0x08),
		1536: int(0x18),
		2048: int(0x28),
		4096: int(0x0C),
	}
	var PAC map[int]int = map[int]int{
		8:  0,
		16: 1,
		32: 2,
		64: 3,
	}
	var anchor_role int
	if anchor.role == "Master" {
		anchor_role = 1
	} else if anchor.role == "Slave" {
		anchor_role = 0
	}
	RTLS_CMD_SET_CFG_CCP := ReportsAndMessages.Build_RTLS_CMD_SET_CFG_CCP(
		int(anchor_role),
		int(rf_config["chnum"].(float64)),
		int(PRF[int(rf_config["prf"].(float64))]),
		int(DATARATE[rf_config["datarate"].(float64)]),
		int(rf_config["preamblecode"].(float64)),
		int(PREAMBLE_LEN[int(rf_config["preamblelen"].(float64))]),
		int(PAC[int(rf_config["pac"].(float64))]),
		int(rf_config["nsfd"].(float64)),
		int(anchor.adrx),
		int(anchor.adtx),
		int(rf_config["diagnostic"].(float64)),
		int(rf_config["lag"].(float64)))

	if anchor.connection != nil {
		anchor.connection.Write(RTLS_CMD_SET_CFG_CCP)
		Logger.Logger("SUCCESS : SetRfConfig on the anchor "+anchor.ip, nil)
		if server_connection != nil {
			MessageToServer(map[string]interface{}{"action": "Success", "data": "Success: SetRfConfig on the anchor " + anchor.ip}, server_connection)
		}
	}
}

func StartSpam(anchor *anchor_struct, server_connection *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : AnchorStartSpam", err)
			if server_connection != nil {
				MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: AnchorStartSpam"}, server_connection)
			}
		}
	}()
	if anchor.connection != nil {
		anchor.connection.Write(ReportsAndMessages.Build_RTLS_START_REQ(1))
		Logger.Logger("SUCCESS : AnchorStartSpam "+anchor.ip, nil)
		if server_connection != nil {
			MessageToServer(map[string]interface{}{"action": "Success", "data": "Sucess: AnchorStartSpam " + anchor.ip}, server_connection)
		}
	}
}

func StopSpam(anchor *anchor_struct, server_connection *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : AnchorStopSpam", err)
			if server_connection != nil {
				MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: AnchorStopSpam"}, server_connection)
			}
		}
	}()
	if anchor.connection != nil {
		anchor.connection.Write(ReportsAndMessages.Build_RTLS_START_REQ(0))
		Logger.Logger("SUCCESS : AnchorStopSpam "+anchor.ip, nil)
		if server_connection != nil {
			MessageToServer(map[string]interface{}{"action": "Success", "data": "Sucess: AnchorStopSpam " + anchor.ip}, server_connection)
		}
	}
}

func MessageToServer(map_message map[string]interface{}, server_connection *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : anchor message to server", err)
		}
	}()
	if server_connection != nil {
		json_message, _ := json.Marshal(map_message)
		server_connection.WriteMessage(websocket.TextMessage, json_message)
		Logger.Logger("SUCCESS : anchor message to server: "+string(json_message), nil)
	}
}

func CheckAnchors(apikey string, name string, clientid string, roomid string, organization string, independent_flag string, connect_math_flag string, start_spam_flag bool, rf_config map[string]interface{}, server_connection *websocket.Conn) {
	for {
		func() {
			defer func() {
				if err := recover(); err != nil {
					Logger.Logger("ERROR : CheckAnchors", err)
				}
			}()
			for i := 0; i < len(anchors_array); i++ {
				if anchors_array[i].connection == nil {
					Connect(&anchors_array[i], server_connection)
					SetRfConfig(&anchors_array[i], rf_config, server_connection)
					if start_spam_flag {
						StartSpam(&anchors_array[i], server_connection)
						go Handler(apikey, name, clientid, roomid, organization, independent_flag, connect_math_flag, &(anchors_array[i]), server_connection)
					}
				}
			}
		}()
		time.Sleep(5 * time.Second)
	}
}
