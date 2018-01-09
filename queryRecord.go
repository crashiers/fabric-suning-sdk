package main

import (
	"net/http"
)

func QueryRecord(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST,GET,OPTIONS,PUT,DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Action, Module")
	}

	if r.Method != "POST" {
		OutputJson(w, -1, "requset method is not post", nil)
		return
	}

	if r.PostFormValue("queryType") == "" || r.PostFormValue("queryField") == "" {
		OutputJson(w, -1, "request is null", nil)
		return
	}

	var queryRecordArgs = [][]byte{[]byte("queryRecord"), []byte(r.PostFormValue("queryType")), []byte(r.PostFormValue("queryField"))}

	value, err := base.Query(queryRecordArgs)
	if err != nil {
		OutputJson(w, -1, err.Error(), nil)
		return
	}

	OutputJson(w, 0, "Query record is ok", value)
}
