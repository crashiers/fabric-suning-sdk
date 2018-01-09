package main

import (
	"net/http"
)

func DeleteRecord(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST,GET,OPTIONS,PUT,DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Action, Module")
	}

	if r.Method != "POST" {
		OutputJson(w, -1, "requset method is not post", nil)
		return
	}

	if r.PostFormValue("orgId") == "" || r.PostFormValue("deleteType") == "" || r.PostFormValue("recordId") == "" {
		OutputJson(w, -1, "request is null", nil)
		return
	}

	var deleteRecordArgs = [][]byte{[]byte("deleteRecord"), []byte(r.PostFormValue("orgId")), []byte(r.PostFormValue("deleteType")), []byte(r.PostFormValue("recordId"))}

	value, err := base.Invoke(deleteRecordArgs)
	if err != nil {
		OutputJson(w, -1, err.Error(), nil)
		return
	}

	OutputJson(w, 0, "Delete record is ok", value)
}
