package wrpc_app

import (
	"bytes"
	"encoding/json"
	"github.com/InspirationKlab/wrpc"
	"github.com/InspirationKlab/wrpc/example"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

func WsAppEntry( /*marker:deps*/ third bytes.Buffer, context example.AppContext /*marker:deps*/) func(writer http.ResponseWriter, request *http.Request) {
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
					return
				}
				var parsed wrpc.RequestBase
				json.Unmarshal(message, &parsed)
				switch parsed.Command {
				/*marker:switch*/

				case "StreamMessages":
					go func(request wrpc.RequestBase) {
						typedRequestArg, err := wrpc.AsTyped[int](&request)
						if err != nil {
							outChannel <- wrpc.Response{
								Id:    request.Id,
								Error: "Error in parsing json: " + err.Error(),
							}
							return
						}
						result := example.StreamMessages(typedRequestArg, context, context, third)
						for rEntry := range result {
							outChannel <- wrpc.Response{
								Id:   request.Id,
								Data: rEntry,
							}
						}

					}(parsed)
					/*marker:switch*/
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
