/*
Copyright 2016 The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/platform9/fission"
)

type API struct {
	FunctionStore
	HTTPTriggerStore
	EnvironmentStore
}

func (api *API) respondWithSuccess(w http.ResponseWriter, resp []byte) {
	_, err := w.Write(resp)
	if err != nil {
		// this will probably fail too, but try anyway
		api.respondWithError(w, err)
	}
}

func (api *API) respondWithError(w http.ResponseWriter, err error) {
	var code int
	var msg string
	debug.PrintStack()
	fe, ok := err.(fission.Error)
	if ok {
		msg = fe.Message
		switch fe.Code {
		case fission.ErrorNotFound:
			code = 404
		case fission.ErrorInvalidArgument:
			code = 400
		case fission.ErrorNoSpace:
			code = 500
		case fission.ErrorNotAuthorized:
			code = 403
		default:
			code = 500
		}
	} else {
		code = 500
		msg = err.Error()
	}
	log.Printf("Error: %v: %v", code, msg)
	http.Error(w, msg, code)
}

func (api *API) HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Fission API")
}

func (api *API) serve(port int) {
	r := mux.NewRouter()
	r.HandleFunc("/", api.HomeHandler)

	r.HandleFunc("/functions", api.FunctionApiList).Methods("GET")
	r.HandleFunc("/functions", api.FunctionApiCreate).Methods("POST")
	r.HandleFunc("/functions/{function}", api.FunctionApiGet).Methods("GET")
	r.HandleFunc("/functions/{function}", api.FunctionApiUpdate).Methods("PUT")
	r.HandleFunc("/functions/{function}", api.FunctionApiDelete).Methods("DELETE")

	r.HandleFunc("/triggers/http", api.HTTPTriggerApiList).Methods("GET")
	r.HandleFunc("/triggers/http", api.HTTPTriggerApiCreate).Methods("POST")
	r.HandleFunc("/triggers/http/{httpTrigger}", api.HTTPTriggerApiGet).Methods("GET")
	r.HandleFunc("/triggers/http/{httpTrigger}", api.HTTPTriggerApiUpdate).Methods("PUT")
	r.HandleFunc("/triggers/http/{httpTrigger}", api.HTTPTriggerApiDelete).Methods("DELETE")

	// r.HandleFunc("/environments", api.EnvironmentApiList).Methods("GET")
	// r.HandleFunc("/environments", api.EnvironmentApiCreate).Methods("POST")
	// r.HandleFunc("/environments/{environment}", api.EnvironmentApiGet).Methods("GET")
	// r.HandleFunc("/environments/{environment}", api.EnvironmentApiUpdate).Methods("PUT")
	// r.HandleFunc("/environments/{environment}", api.EnvironmentApiDelete).Methods("DELETE")

	address := fmt.Sprintf(":%v", port)

	log.WithFields(log.Fields{"port": port}).Info("Server started")
	log.Fatal(http.ListenAndServe(address, handlers.LoggingHandler(os.Stdout, r)))
}
