package main

import (
	"fmt"
	"strconv"
	"testing"
)

var count int = 20000

func TestLoad(t *testing.T) {

	var createOrgs = [][][]byte{[][]byte{[]byte("createOrg"), []byte("1"), []byte("performancetesting1")},
		[][]byte{[]byte("createOrg"), []byte("2"), []byte("performancetesting2")},
		[][]byte{[]byte("createOrg"), []byte("3"), []byte("performancetesting3")},
		[][]byte{[]byte("createOrg"), []byte("4"), []byte("performancetesting4")},
		[][]byte{[]byte("createOrg"), []byte("5"), []byte("performancetesting5")}}

	thread := len(createOrgs)

	var chans = make(chan bool, thread)

	for i := 0; i < thread; i++ {
		_, err := base.Invoke(createOrgs[i])
		if err != nil {
			return
		}
		go func() {
			for j := 0; j < count; j++ {
				var submitRecordArgs = [][]byte{[]byte("submitRecord"), []byte(strconv.Itoa(i)), []byte(strconv.Itoa(j)),
					[]byte("111"), []byte("clientName"), []byte("a"),
					[]byte("low"), []byte("negativeInfo")}

				_, err := base.Invoke(submitRecordArgs)
				if err != nil {
					fmt.Println(err)
				}
			}
			chans <- true
		}()
	}

	var tcount = 0
	for {
		<-chans
		tcount++
		if tcount == thread {
			break
		}
	}

}
