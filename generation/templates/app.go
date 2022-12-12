package templates

import "net/http"

func WsAppEntry() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		/*marker:content*/
		/*marker:entries:for:entry*/
		/*marker:entry*/
		/*marker:entries:end*/
	}
}
