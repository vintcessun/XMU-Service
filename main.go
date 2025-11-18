package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/vintcessun/XMU-Service/api"
	"github.com/vintcessun/XMU-Service/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func easySendMessage(conn *websocket.Conn, message string) {
	if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		log.Printf("发送消息失败: %v", err)
		return
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("升级连接失败: %v", err)
		return
	}
	defer conn.Close()

	log.Println("客户端已连接")

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("读取消息失败: %v", err)
			return
		}

		log.Printf("收到消息: %s", msg)
		switch msgType {
		case websocket.TextMessage:
			msgs := strings.Split(string(msg), " ")
			switch msgs[0] {
			case "login_lnt_qr":
				go func() {
					lntClient := api.LntServiceQr{}
					err := lntClient.GetInfo()
					if err != nil {
						easySendMessage(conn, "Error "+err.Error())
						return
					}
					easySendMessage(conn, "QrCodeId "+lntClient.QrcodeId)
					for {
						state, err := lntClient.GetQrState()
						if err != nil {
							easySendMessage(conn, "Error "+err.Error())
							return
						}

						if state == "0" {
							//等待中
						} else if state == "1" {
							//请求成功
							break
						} else if state == "2" {
							//已扫描二维码
						} else if state == "3" {
							//二维码已失效
							easySendMessage(conn, "Error QrCodeExpired")
							return
						}
					}

					err = lntClient.Finish()
					if err != nil {
						easySendMessage(conn, "Error "+err.Error())
						return
					}

					easySendMessage(conn, "Session "+lntClient.Session)
				}()
			case "login_lnt_password":
				go func() {
					if len(msgs) != 3 {
						easySendMessage(conn, "Error 参数错误")
						return
					}
					lntClient := api.LntServicePassword{Username: strings.TrimSpace(msgs[1]), Password: strings.TrimSpace(msgs[2])}
					err := lntClient.Login()
					if err != nil {
						easySendMessage(conn, "Error "+err.Error())
						return
					}
					easySendMessage(conn, "Session "+lntClient.Session)
				}()
			case "ping":
				easySendMessage(conn, "pong")
			case "profile_lnt":
				go func() {
					if len(msgs) != 2 {
						easySendMessage(conn, "Error 参数错误")
						return
					}
					session := strings.TrimSpace(msgs[1])
					result, err := api.GetProfile(session)
					if err != nil {
						easySendMessage(conn, "Error "+err.Error())
						return
					}
					data, err := utils.MarshalJSON[*api.ProfileResponse](result)
					if err != nil {
						easySendMessage(conn, "Error "+err.Error())
						return
					}
					easySendMessage(conn, "Profile "+data)
				}()
			case "login_zyh365":
				go func() {
					if len(msgs) != 3 {
						easySendMessage(conn, "Error 参数错误")
						return
					}
					username := strings.TrimSpace(msgs[1])
					password := strings.TrimSpace(msgs[2])
					zyh365Client := api.Zyh365ServicePassword{Username: username, Password: password}
					err := zyh365Client.Login()
					if err != nil {
						easySendMessage(conn, "Error "+err.Error())
						return
					}
					easySendMessage(conn, "Token "+zyh365Client.Token)
				}()
			case "hours_zyh365":
				go func() {
					if len(msgs) != 3 {
						easySendMessage(conn, "Error 参数错误")
						return
					}
					Token := strings.TrimSpace(msgs[1])
					userName := strings.TrimSpace(msgs[2])
					service := api.Zyh365ServiceHours{}
					err := service.GetHours(Token, userName)
					if err != nil {
						easySendMessage(conn, "Error "+err.Error())
						return
					}
					data, err := utils.MarshalJSON[*api.Zyh365ServiceHours](&service)
					if err != nil {
						easySendMessage(conn, "Error "+err.Error())
						return
					}
					easySendMessage(conn, "Hours "+data)
				}()
			}
		case websocket.BinaryMessage:
			easySendMessage(conn, "The command is not allowed")
		}

	}
}

func main() {
	http.HandleFunc("/ws", wsHandler)

	log.Println("WebSocket服务启动，监听端口 8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
