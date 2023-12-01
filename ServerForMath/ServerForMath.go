package ServerForMath

import (
	"NODE/Logger"
	"encoding/json"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type client_struct struct {
	mutex       sync.Mutex
	connection  *websocket.Conn
	client_type string
}

var clients []client_struct
var anchors_array []map[string]interface{}
var ref_tag_config map[string]interface{}

func RoomAndReftagConfig(anchors []map[string]interface{}, ref_tag map[string]interface{}) {
	anchors_array = anchors
	ref_tag_config = ref_tag
	for i := 0; i < len(clients); i++ {
		if clients[i].client_type == "math" && clients[i].connection != nil {
			MessageToMath(map[string]interface{}{"action": "RoomConfig", "data": map[string]interface{}{"clientid": "clientid", "organization": "clientid", "roomid": "roomid", "roomname": "roomname", "anchors": anchors_array, "ref_tag_config": ref_tag_config}})

		}
	}
}

func GenerateApikey() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789!№;%:?*()-_+=")
	apikey := make([]rune, 30)
	rand.Seed(time.Now().UnixNano())
	for i := range apikey {
		apikey[i] = letters[rand.Intn(len(letters))]
	}
	return string(apikey)
}

func Receiver(connection *websocket.Conn, server_connection *websocket.Conn) {
	math_apikey := ""
	break_flag := false
	for {
		if break_flag {
			return
		}
		func() {
			defer func() {
				if err := recover(); err != nil {
					Logger.Logger("ERROR : ServerForMath Receiver", err)
					if err.(string) == "repeated read on failed websocket connection" {
						break_flag = true
						for i := 0; i < len(clients); i++ {
							if clients[i].connection != nil {
								if clients[i].connection.LocalAddr() == connection.LocalAddr() {
									clients[i].connection.Close()
									var nil_websocket *websocket.Conn
									clients[i].connection = nil_websocket
									clients[i].client_type = ""
								}
							}
						}
					}
				}
			}()

			_, message, _ := connection.ReadMessage()
			var message_map map[string]interface{}
			json.Unmarshal(message, &message_map)
			Logger.Logger("SUCCESS : Message from node's client: "+string(message), nil)
			if message_map["action"] == "Login" && message_map["login"] == "mathLogin" && message_map["password"] == "%wPp7VO6k7ump{BP4mu2rm4w?p|J5N%P" {
				for i := 0; i < len(clients); i++ {
					if clients[i].connection == connection {
						clients[i].client_type = "math"
					}
				}
				math_apikey = GenerateApikey()
				MessageToMath(map[string]interface{}{"action": "Login", "data": map[string]interface{}{"apikey": math_apikey}})
				if anchors_array != nil && ref_tag_config != nil {
					MessageToMath(map[string]interface{}{"action": "RoomConfig", "data": map[string]interface{}{"clientid": "clientid", "organization": "clientid", "roomid": "roomid", "roomname": "roomname", "anchors": anchors_array, "ref_tag_config": ref_tag_config}})
				}
				if server_connection != nil {
					json_message_for_server, _ := json.Marshal(map[string]interface{}{"action": "Success", "data": "math connected"})
					server_connection.WriteMessage(websocket.TextMessage, json_message_for_server)
					Logger.Logger("SUCCESS : Message to server: "+string(json_message_for_server), nil)
				}
			} else if message_map["apikey"] == math_apikey {
				if server_connection != nil {
					server_connection.WriteMessage(websocket.TextMessage, message)
					Logger.Logger("SUCCESS : Message from math to server: "+string(message), nil)
				}
				for i := 0; i < len(clients); i++ {
					if clients[i].connection != connection && clients[i].connection != nil {
						clients[i].mutex.Lock()
						clients[i].connection.WriteMessage(websocket.TextMessage, message)
						clients[i].mutex.Unlock()
						Logger.Logger("SUCCESS : Message from math to client: "+string(message), nil)
					}
				}
			}
		}()
	}
}

func StartServer(node_server_ip string, node_server_port string, server_connection *websocket.Conn) {
	var upgrader = websocket.Upgrader{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			Logger.Logger("ERROR : Node's server connection", err)
			if server_connection != nil {
				json_message_for_server, _ := json.Marshal(map[string]interface{}{"action": "Error", "data": "Node's server connection"})
				server_connection.WriteMessage(websocket.TextMessage, json_message_for_server)
				Logger.Logger("SUCCESS: Message to server: "+string(json_message_for_server), nil)
			}
		}
		Logger.Logger("SUCCESS : New websocket connection for node", nil)
		if server_connection != nil {
			json_message_for_server, _ := json.Marshal(map[string]interface{}{"action": "Success", "data": "New websocket connection for node"})
			server_connection.WriteMessage(websocket.TextMessage, json_message_for_server)
			Logger.Logger("SUCCESS: Message to server: "+string(json_message_for_server), nil)
		}
		clients = append(clients, client_struct{connection: connection, client_type: "client"})
		go Receiver(connection, server_connection)
	})
	http.ListenAndServe(node_server_ip+":"+node_server_port, nil)
}

func MessageToMath(map_message map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Logger("ERROR : MessageToMath", err)
		}
	}()

	for i := 0; i < len(clients); i++ {
		if clients[i].client_type == "math" && clients[i].connection != nil {
			json_message, _ := json.Marshal(map_message)
			clients[i].mutex.Lock()
			clients[i].connection.WriteMessage(websocket.TextMessage, json_message)
			clients[i].mutex.Unlock()
			Logger.Logger("SUCCESS : Message to math: "+string(json_message), nil)
		}
	}
}
