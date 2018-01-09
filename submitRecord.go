package main

import (
	"net/http"
)

func SubmitRecord(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST,GET,OPTIONS,PUT,DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Action, Module")
	}

	if r.Method != "POST" {
		OutputJson(w, -1, "requset method is not post", nil)
		return
	}

	if r.PostFormValue("orgId") == "" || r.PostFormValue("recordId") == "" || r.PostFormValue("clientId") == "" ||
		r.PostFormValue("clientName") == "" || r.PostFormValue("negativeType") == "" || r.PostFormValue("negativeSeverity") == "" ||
		r.PostFormValue("negativeInfo") == "" {
		OutputJson(w, -1, "request is null", nil)
		return
	}

	var submitRecordArgs = [][]byte{[]byte("submitRecord"), []byte(r.PostFormValue("orgId")), []byte(r.PostFormValue("recordId")),
		[]byte(r.PostFormValue("clientId")), []byte(r.PostFormValue("clientName")), []byte(r.PostFormValue("negativeType")),
		[]byte(r.PostFormValue("negativeSeverity")), []byte(r.PostFormValue("negativeInfo"))}

	value, err := base.Invoke(submitRecordArgs)
	if err != nil {
		OutputJson(w, -1, err.Error(), nil)
		return
	}

	OutputJson(w, 0, "Submit record is ok", value)
}
