package wrpc_app

import (
	"blacksec.com/wrpc/v2"
	"blacksec.com/wrpc/v2/example"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

func WsAppEntry( /* wrpc-app-deps:begin */ context example.AppContext /* wrpc-app-deps:end */) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		upgrade := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
			return true
		}}
		connection, err := upgrade.Upgrade(writer, request, nil)
		if err != nil {
			http.Error(writer, err.Error(), 500)
		}
		outChannel := make(chan wrpc.Response)
		defer connection.Close()

		go func() {
			for err == nil {
				mt, message, err := connection.ReadMessage()
				if err != nil || mt == websocket.CloseMessage {
					return // Выходим из цикла, если клиент пытается закрыть соединение или связь прервана
				}
				var parsed wrpc.RequestBase
				json.Unmarshal(message, &parsed)
				switch parsed.Command {
				// wrpc-app-list

				case "StreamMessages":
					go func(request wrpc.RequestBase) {
						typedArg := int(parsed.Args.(float64))
						result := example.StreamMessages(typedArg, context)
						for {
							select {
							case value := <-result:
								outChannel <- wrpc.Response{
									Id:   request.Id,
									Data: value,
								}
							}
						}
					}(parsed)
					break

				case "Ping":
					go func(request wrpc.RequestBase) {
						typedArg := float64(parsed.Args.(float64))
						result := example.Ping(typedArg)
						outChannel <- wrpc.Response{
							Id:   request.Id,
							Data: result,
						}
					}(parsed)
					break

				}
			}

		}()
		for {
			select {
			case output := <-outChannel:
				err = connection.WriteJSON(output)
				if err != nil {
					log.Printf("Error in writing message to websocket, %v", err)
					return
				}
			}
		}
	}
}
