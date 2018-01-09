package main

import (
	"net/http"
)

func Transfer(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST,GET,OPTIONS,PUT,DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Action, Module")
	}

	if r.Method != "POST" {
		OutputJson(w, -1, "requset method is not post", nil)
		return
	}

	if r.PostFormValue("fromId") == "" || r.PostFormValue("toId") == "" || r.PostFormValue("creditNumber") == "" {
		OutputJson(w, -1, "request is null", nil)
		return
	}

	var transferArgs = [][]byte{[]byte("transfer"), []byte(r.PostFormValue("fromId")), []byte(r.PostFormValue("toId")), []byte(r.PostFormValue("creditNumber"))}

	value, err := base.Invoke(transferArgs)
	if err != nil {
		OutputJson(w, -1, err.Error(), nil)
		return
	}

	OutputJson(w, 0, "Transfer is ok", value)
}
