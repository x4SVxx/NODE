package main

import (
	"NODE/Anchor"
	"NODE/Logger"
	"NODE/ReadAndSetNodeConfig"
	"NODE/ServerForMath"

	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var server_ip, server_port, login, password, roomid, independent_flag, connect_math_flag, node_server_ip, node_server_port, ref_tag_config = ReadAndSetNodeConfig.ReadAndSetNodeConfig()
var apikey, clientid, name, roomname, organization string
var login_flag, config_flag, start_spam_flag bool = false, false, false
var rf_config = map[string]interface{}{}

type server_struct struct {
	mutex      sync.Mutex
	connection *websocket.Conn
}

var server server_struct

func main() {
	if independent_flag == "true" {
		go ServerForMath.StartServer(node_server_ip, node_server_port, server.connection)
		time.Sleep(1 * time.Second)

		// cmd := exec.Command("python", "Math/test_con.py")
		// out, err := cmd.Output()
		// fmt.Println(string(out))
		// if err != nil {
		// 	fmt.Println("err")
		// }

		var anchors_reader []interface{}
		var rf_config_reader []interface{}
		config_file, _ := os.ReadFile("Config.json")
		rf_config_file, _ := os.ReadFile("RfConfig.json")
		json.Unmarshal(config_file, &anchors_reader)
		json.Unmarshal(rf_config_file, &rf_config_reader)
		login_flag = true
		SetConfig(map[string]interface{}{"anchors": anchors_reader, "rf_config": rf_config_reader})
		time.Sleep(2 * time.Second)
		StartSpam()
		for {
			var name string
			fmt.Scanf("%s\n", &name)
			if name == "Stop" {
				StopSpam()
			}
			if name == "Start" {
				StartSpam()
			}
		}
	}

	if independent_flag == "false" {
		for {
			func() {
				defer func() {
					if err := recover(); err != nil {
						Logger.Logger("ERROR : main - server handler", err)
					}
				}()
				URL := url.URL{Scheme: "ws", Host: server_ip + ":" + server_port}
				var error_server_connection error
				server.connection, _, error_server_connection = websocket.DefaultDialer.Dial(URL.String(), nil)
				if error_server_connection != nil {
					Logger.Logger("ERROR : main - server connection", error_server_connection)
				} else {
					Logger.Logger("SUCCESS : node connected to the server "+string(server_ip)+":"+string(server_port), nil)
					MessageToServer(map[string]interface{}{"action": "Login", "login": login, "password": password, "roomid": roomid})
					if connect_math_flag == "true" {
						go ServerForMath.StartServer(node_server_ip, node_server_port, server.connection)
					}
					break_main_receiver_point := false
					for {
						if break_main_receiver_point {
							break
						}
						func() {
							defer func() {
								if err := recover(); err != nil {
									Logger.Logger("ERROR : main receiver", err)
									if err.(string) == "repeated read on failed websocket connection" {
										break_main_receiver_point = true
									}
									if start_spam_flag {
										StopSpam()
									}
								}
							}()
							_, message, error_read_message_from_server := server.connection.ReadMessage()
							if error_read_message_from_server != nil {
								Logger.Logger("ERROR : message from server", error_read_message_from_server)
							} else {
								Logger.Logger("SUCCESS : message from server "+string(message), nil)
								var message_map map[string]interface{}
								error_unmarshal_json := json.Unmarshal(message, &message_map)
								if error_unmarshal_json != nil {
									Logger.Logger("ERROR : Unmarshal message from server", error_unmarshal_json)
								} else {
									if message_map["action"] == "Login" && message_map["status"] == "true" {
										Login(message_map)
									}
									if message_map["action"] == "SetConfig" && message_map["status"] == "true" {
										SetConfig(message_map)
									}
									if message_map["action"] == "Start" && message_map["status"] == "true" {
										StartSpam()
									}
									if message_map["action"] == "Stop" && message_map["status"] == "true" {
										StopSpam()
									}
								}
							}
						}()
					}
				}
			}()
		}
	}
}

func Login(message_map map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			login_flag = false
			Logger.Logger("ERROR : Login", err)
			MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: Login"})
		}
	}()
	apikey = string(message_map["data"].(map[string]interface{})["apikey"].(string))
	clientid = string(message_map["data"].(map[string]interface{})["clientid"].(string))
	name = string(message_map["data"].(map[string]interface{})["name"].(string))
	roomname = string(message_map["data"].(map[string]interface{})["roomname"].(string))
	organization = string(message_map["data"].(map[string]interface{})["organization"].(string))

	if len(strings.TrimSpace(apikey)) == 0 ||
		len(strings.TrimSpace(clientid)) == 0 ||
		len(strings.TrimSpace(name)) == 0 ||
		len(strings.TrimSpace(roomname)) == 0 ||
		len(strings.TrimSpace(organization)) == 0 {
		MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: Login - empty data"})
		login_flag = false
	} else {
		MessageToServer(map[string]interface{}{"action": "Success", "data": "Success: Login"})
		login_flag = true
	}
}

func SetConfig(message_map map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : main SetConfig", err)
			MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: main SetConfig"})
		}
	}()
	if !login_flag {
		MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: SetConfig - need autorization"})
	} else {
		if start_spam_flag {
			StopSpam()
		}

		Anchor.DisConnectAnchors(server.connection)
		Anchor.ClearAnchors()
		time.Sleep(1 * time.Second)
		for i := 0; i < len(message_map["anchors"].([]interface{})); i++ {
			Anchor.CreateAnchor(message_map["anchors"].([]interface{})[i].(map[string]interface{}))
		}
		Anchor.ConnectAnchors(server.connection)
		rf_config["chnum"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["chnum"].(float64)
		rf_config["prf"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["prf"].(float64)
		rf_config["datarate"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["datarate"].(float64)
		rf_config["preamblecode"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["preamblecode"].(float64)
		rf_config["preamblelen"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["preamblelen"].(float64)
		rf_config["pac"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["pac"].(float64)
		rf_config["nsfd"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["nsfd"].(float64)
		rf_config["diagnostic"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["diagnostic"].(float64)
		rf_config["lag"] = message_map["rf_config"].([]interface{})[0].(map[string]interface{})["lag"].(float64)

		Anchor.SetRfConfigAnchors(rf_config, server.connection)

		config_flag = true

		Anchor.SetRoomConfigToMath(ref_tag_config, apikey, name, clientid, roomid, organization, roomname, independent_flag, connect_math_flag, server.connection)
	}
}

func StartSpam() {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : main StartSpam", err)
			MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: main StartSpam"})
		}
	}()
	if !config_flag {
		MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: StartSpam - need config"})
	} else {
		if !start_spam_flag {
			start_spam_flag = true
			Anchor.StartSpamAnchors(apikey, name, clientid, roomid, organization, roomname, independent_flag, connect_math_flag, start_spam_flag, rf_config, server.connection)
		}
	}
}

func StopSpam() {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : main StopSpam", err)
			MessageToServer(map[string]interface{}{"action": "Error", "data": "Error: Error main StopSpam"})
		}
	}()
	if start_spam_flag {
		start_spam_flag = false
		Anchor.StopSpamAnchors(server.connection)
	}
}

func MessageToServer(map_message map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : main - MessageToServer", err)
		}
	}()
	if server.connection != nil {
		json_message, _ := json.Marshal(map_message)
		server.mutex.Lock()
		server.connection.WriteMessage(websocket.TextMessage, json_message)
		server.mutex.Unlock()
		Logger.Logger("SUCCESS : main - Message to server: "+string(json_message), nil)
	}
}
