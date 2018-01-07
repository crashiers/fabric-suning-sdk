package main

import (
	"net/http"
)

func CreatOrg(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST,GET,OPTIONS,PUT,DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Action, Module")
	}

	if r.Method != "POST" {
		OutputJson(w, -1, "requset method is not post", nil)
		return
	}

	if r.PostFormValue("orgId") == "" || r.PostFormValue("orgName") == "" {
		OutputJson(w, -1, "request is null", nil)
		return
	}

	var args []string
	args = append(args, "invoke")
	args = append(args, "createOrg")
	args = append(args, r.PostFormValue("orgId"))
	args = append(args, r.PostFormValue("orgName"))

	//value, err := base.Invoke(args)
	//	if err != nil {
	//		OutputJson(w, -1, err.Error(), nil)
	//		return
	//	}

	//OutputJson(w, 0, "Create organazation is ok", value)
}
