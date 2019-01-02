//
// Copyright (c) 2018 Tencent
//
// Copyright (c) 2018 Dell Inc.
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
)

func LoadRestRoutes() http.Handler {
	r := mux.NewRouter()

	// Ping Resource
	r.HandleFunc(clients.ApiPingRoute, pingHandler).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(clients.ApiConfigRoute, configHandler).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(clients.ApiMetricsRoute, metricsHandler).Methods(http.MethodGet)

	b := r.PathPrefix(clients.ApiBase).Subrouter()

	// Info
	b.HandleFunc("/info/{name}", replyInfo).Methods(http.MethodGet)

	// Flush reload schedules
	b.HandleFunc("/flush", replyFlushScheduler).Methods(http.MethodGet)

	// Callback
	r.HandleFunc(clients.ApiCallbackRoute, addCallbackAlert).Methods(http.MethodPost)
	r.HandleFunc(clients.ApiCallbackRoute, updateCallbackAlert).Methods(http.MethodPut)
	r.HandleFunc(clients.ApiCallbackRoute, removeCallbackAlert).Methods(http.MethodDelete)

	return r
}

// Test if the service is working
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
}

func configHandler(w http.ResponseWriter, _ *http.Request) {
	encode(Configuration, w)
}

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	var t internal.Telemetry

	// The micro-service is to be considered the System Of Record (SOR) in terms of accurate information.
	// Fetch metrics for the scheduler service.
	var rtm runtime.MemStats

	// Read full memory stats
	runtime.ReadMemStats(&rtm)

	// Miscellaneous memory stats
	t.Alloc = rtm.Alloc
	t.TotalAlloc = rtm.TotalAlloc
	t.Sys = rtm.Sys
	t.Mallocs = rtm.Mallocs
	t.Frees = rtm.Frees

	// Live objects = Mallocs - Frees
	t.LiveObjects = t.Mallocs - t.Frees

	encode(t, w)

	return
}

func replyInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(ContentTypeKey, ContentTypeJsonValue)

	vars := mux.Vars(r)
	name := vars["name"]
	schedule, err := queryScheduleByName(name)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("read info request error %s", err.Error()))
		http.Error(w, "Schedule/Event not found", http.StatusNotFound)
		return
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(schedule); err != nil {
		LoggingClient.Error(fmt.Sprintf("Error encoding the data: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func replyFlushScheduler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(ContentTypeKey, ContentTypeJsonValue)

	err := AddSchedulers()
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error reloading new schedules, scheduleEvents,  or addressables: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	str := `{"flush" : "success"}`
	io.WriteString(w, str)
}

func addCallbackAlert(rw http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("read request body error : %s", err.Error()))
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	callbackAlert := models.CallbackAlert{}
	if err := json.Unmarshal(data, &callbackAlert); err != nil {
		LoggingClient.Error(fmt.Sprintf("failed to parse callback alert : %s", err.Error()))
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	switch callbackAlert.ActionType {

	case models.SCHEDULE:

		schedule, err := querySchedule(callbackAlert.Id)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("query schedule error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		err = addSchedule(schedule)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("add schedule error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		break

	case models.SCHEDULEEVENT:

		scheduleEvent, err := queryScheduleEvent(callbackAlert.Id)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("query schedule event error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := addScheduleEvent(scheduleEvent); err != nil {
			LoggingClient.Error(fmt.Sprintf("add schedule event error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		break

	default:
		LoggingClient.Error(fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType))
		http.Error(rw, fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType), http.StatusBadRequest)
		break
	}
}

func updateCallbackAlert(rw http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("reading the http request body error : %s", err.Error()))
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	callbackAlert := models.CallbackAlert{}
	if err := json.Unmarshal(data, &callbackAlert); err != nil {
		LoggingClient.Error(fmt.Sprintf("failed to parse callback alert : %s", err.Error()))
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	switch callbackAlert.ActionType {

	case models.SCHEDULE:

		schedule, err := querySchedule(callbackAlert.Id)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("query schedule error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		err = updateSchedule(schedule)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("update schedule error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		break

	case models.SCHEDULEEVENT:

		scheduleEvent, err := queryScheduleEvent(callbackAlert.Id)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("query schedule event error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := updateScheduleEvent(scheduleEvent); err != nil {
			LoggingClient.Error(fmt.Sprintf("query schedule event error :%s ", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		break

	default:
		LoggingClient.Error(fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType))
		http.Error(rw, fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType), http.StatusBadRequest)
		break
	}
}

func removeCallbackAlert(rw http.ResponseWriter, r *http.Request) {
	//here we need the action type, so request the callback alert json
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("reading the http request body error : %s", err.Error()))
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	callbackAlert := models.CallbackAlert{}
	if err := json.Unmarshal(data, &callbackAlert); err != nil {
		LoggingClient.Error(fmt.Sprintf("failed to parse callback alert : %s", err.Error()))
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	switch callbackAlert.ActionType {
	case models.SCHEDULE:
		if err := removeSchedule(callbackAlert.Id); err != nil {
			LoggingClient.Error(fmt.Sprintf("remove schedule error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusOK)
		}
		break

	case models.SCHEDULEEVENT:
		if err := removeScheduleEvent(callbackAlert.Id); err != nil {
			LoggingClient.Error(fmt.Sprintf("remove schedule event error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusOK)
		}
		break

	default:
		LoggingClient.Error(fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType))
		http.Error(rw, fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType), http.StatusBadRequest)
		break
	}
}

// Helper function for encoding things for returning from REST calls
func encode(i interface{}, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	err := enc.Encode(i)
	// Problems encoding
	if err != nil {
		LoggingClient.Error("Error encoding the data: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
