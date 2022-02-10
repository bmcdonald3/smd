// MIT License
// 
// (C) Copyright [2022] Hewlett Packard Enterprise Development LP
// 
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package service_reservations

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"github.com/sirupsen/logrus"
)


const (
	hsmReservationPath = "/hsm/v2/locks/service/reservations"
	hsmReservationReleasePath = "/hsm/v2/locks/service/reservations/release"
	hsmReservationRenewPath = "/hsm/v2/locks/service/reservations/renew"
)

var prod = &Production{}
var smServer *httptest.Server
var initDone = false
var failAquire = false
var logger = logrus.New()

//Storage of our fake reservations

var resMap map[string]*ReservationCreateSuccessResponse


/////////////////////////////////////////////////////////////////////////////
// Funcs to simulate HSM APIs for reservations.
/////////////////////////////////////////////////////////////////////////////

func smReservationHandler(w http.ResponseWriter, r *http.Request) {
	fname := "smReservationHandler()"
	var jdata ReservationCreateParameters
	var rsp ReservationCreateResponse

	body,_ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body,&jdata)
	if (err != nil) {
		logger.Errorf("%s: Error unmarshalling req data: %v",fname,err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	now := time.Now()
	for ix,comp := range(jdata.ID) {
		if (failAquire && (ix == 0)) {
			rsp.Failure = append(rsp.Failure,FailureResponse{ID: comp,
							Reason: "Forced Failure"})
		} else {
			lres := ReservationCreateSuccessResponse{ID: comp,
				        ReservationKey: fmt.Sprintf("RSVKey_%d",ix),
						DeputyKey: fmt.Sprintf("DEPKey_%d",ix),
						ExpirationTime: now.Format(time.RFC3339)}
			resMap[comp] = &lres
			rsp.Success = append(rsp.Success,lres)
		}
	}

	//TODO: how to simulate failure components?

	ba,baerr := json.Marshal(&rsp)
	if (baerr != nil) {
		logger.Errorf("%s: Error marshalling response data: %v",fname,baerr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

func smReservationRenewHandler(w http.ResponseWriter, r *http.Request) {
	var inData ReservationRenewalParameters
	var retData ReservationReleaseRenewResponse
	fname := "smReservationRenewHandler()"

	body,_ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body,&inData)
	if (err != nil) {
		logger.Errorf("%s: Error unmarshalling req data: %v",fname,err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Just copy gozintas into gozoutas

	for _,key := range(inData.ReservationKeys) {
		retData.Success.ComponentIDs = append(retData.Success.ComponentIDs,
			key.ID)
	}

	retData.Counts.Success = len(inData.ReservationKeys)
	retData.Counts.Failure = 0
	retData.Counts.Total   = retData.Counts.Success + retData.Counts.Failure

	ba,baerr := json.Marshal(&retData)
	if (baerr != nil) {
		logger.Errorf("%s: Error marshalling response data: %v",fname,baerr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

func smReservationReleaseHandler(w http.ResponseWriter, r *http.Request) {
	var relList ReservationReleaseParameters
	var retData ReservationReleaseRenewResponse
	fname := "smReservationReleaseHandler()"

	body,_ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body,&relList)
	if (err != nil) {
		logger.Errorf("%s: Error unmarshalling req data: %v",fname,err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _,comp := range(relList.ReservationKeys) {
		//Check for key existence, present == success, else failure
		_,ok := resMap[comp.ID]
		if (ok) {
			delete(resMap,comp.ID)
			retData.Success.ComponentIDs = append(retData.Success.ComponentIDs,
				comp.ID)
		} else {
			retData.Failure = append(retData.Failure,
					FailureResponse{ID: comp.ID, Reason: "Reservation not found."})
		}
	}
	retData.Counts.Success = len(retData.Success.ComponentIDs)
	retData.Counts.Failure = len(retData.Failure)
	retData.Counts.Total   = retData.Counts.Success + retData.Counts.Failure

	ba,baerr := json.Marshal(&retData)
	if (baerr != nil) {
		logger.Errorf("%s: Error marshalling response data: %v",fname,baerr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

// Insure various stuff is initialized.  Needed since we don't know which
// test will be run when.

func checkInit() {
	if (!initDone) {
		mux := http.NewServeMux()
		mux.HandleFunc(hsmReservationPath,
			http.HandlerFunc(smReservationHandler))
		mux.HandleFunc(hsmReservationReleasePath,
			http.HandlerFunc(smReservationReleaseHandler))
		mux.HandleFunc(hsmReservationRenewPath,
			http.HandlerFunc(smReservationRenewHandler))
		smServer = httptest.NewServer(mux)
		//logger.SetLevel(logrus.TraceLevel)
		prod.InitInstance(smServer.URL,"",1,logger,"RSVTest")
		resMap = make(map[string]*ReservationCreateSuccessResponse)
		initDone = true
	}
}

// Test the rigid reservation model.

func TestAquire(t *testing.T) {
	checkInit()

	xnames := []string{"x0c0s0b0n0","x1c1s1b1n1"}
	err := prod.Aquire(xnames)
	if (err != nil) {
		t.Errorf("Aquire() failed: %v",err)
	}

	ok := prod.Check(xnames)
	if (ok != true) {
		t.Errorf("Check() failed!")
	}

	//Test Check() and Release() (rigid model) with unknown xname

	xn2 := make([]string,len(xnames))
	copy(xn2,xnames)
	xn2 = append(xn2,"x10c7s7b0n0")
	ok = prod.Check(xn2)
	if (ok == true) {
		t.Errorf("Check() should have failed (unk. xname) but didn't.")
	}

	err = prod.Release(xn2)
	if (err == nil) {
		t.Errorf("Release() should have failed (unk. xname), but didn't.")
	}

	//Now do a good Release().

	err = prod.Release(xnames)
	if (err != nil) {
		t.Errorf("Release() failed: %v",err)
	}

	//Test aquire() but force a failure in the server

	failAquire = true
	err = prod.Aquire(xnames)
	if (err == nil) {
		t.Errorf("Aquire() should have failed, did not.")
	}
	failAquire = false

}

// Test the flexible reservation model.

func TestFlexAquire(t *testing.T) {
	checkInit()

	xnames := []string{"x0c0s0b0n0","x1c1s1b1n1"}
	rsv,err := prod.FlexAquire(xnames)
	if (err != nil) {
		t.Errorf("FlexAquire() failed: %v",err)
	}

	for _,rr := range(rsv.Success) {
		ok := false
		for _,xx := range(xnames) {
			if (rr.ID == xx) {
				ok = true
				break
			}
		}
		if (!ok) {
			t.Errorf("Did not match: '%s'",rr.ID)
		}
	}

	rchk,ok := prod.FlexCheck(xnames)
	if (ok != true) {
		t.Errorf("FlexCheck() failed!")
	}

	for _,rr := range(rchk.Success) {
		ok := false
		for _,xx := range(xnames) {
			if (rr.ID == xx) {
				ok = true
				break
			}
		}
		if (!ok) {
			t.Errorf("FlexCheck did not match success: '%s'",rr.ID)
		}
	}

	//Test Check() and Release() (flex model) with unknown xname

	badComp := "x10c7s7b0n0"
	xn2 := make([]string,len(xnames))
	copy(xn2,xnames)
	xn2 = append(xn2,badComp)
	rchk,ok = prod.FlexCheck(xn2)
	if (ok == true) {
		t.Errorf("FlexCheck() returned TRUE (unk. xname) but shouldn't have.")
	}

	//Should see one failure component, others OK.

	glen := len(xn2) - 1
	if ((len(rchk.Success) != glen) || (len(rchk.Failure) != 1)) {
		t.Errorf("FlexCheck() bad success/failure counts: exp: %d/%d, got %d/%d",
			glen,1,len(rchk.Success),len(rchk.Failure))
	}

	for _,rr := range(rchk.Success) {
		ok := false
		for _,xx := range(xnames) {
			if (rr.ID == xx) {
				ok = true
				break
			}
		}
		if (!ok) {
			t.Errorf("FlexCheck(): Did not match success: '%s'",rr.ID)
		}
	}
	if (rchk.Failure[0].ID != badComp) {
		t.Errorf("FlexCheck(): Did not match failure: exp: '%s', got: '%s'",
			badComp,rchk.Failure[0].ID)
	}

	srel,srelerr := prod.FlexRelease(xn2)
	if (srelerr != nil) {
		t.Errorf("FlexRelease() failed, but should't.")
	}
	if (srel.Counts.Total != len(xn2)) {
		t.Errorf("FlexRelease() bad total count, exp: %d, got: %d",
			len(xn2),srel.Counts.Total)
	}
	if (srel.Counts.Success != glen) {
		t.Errorf("FlexRelease() bad success count, exp: %d, got: %d",
			glen,srel.Counts.Success)
	}
	if (srel.Counts.Failure != 1) {
		t.Errorf("FlexRelease() bad failure count, exp: %d, got: %d",
			1,srel.Counts.Failure)
	}
	for _,rr := range(srel.Success.ComponentIDs) {
		ok := false
		for _,xx := range(xnames) {
			if (rr == xx) {
				ok = true
				break
			}
		}
		if (!ok) {
			t.Errorf("FlexRelease(): Did not match success: '%s'",rr)
		}
	}
	if (srel.Failure[0].ID != badComp) {
		t.Errorf("FlexRelease(): Did not match failure: exp: '%s', got: '%s'",
			badComp,srel.Failure[0].ID)
	}


	//Now do a good FlexAquire() and FlexRelease().

	rsv,err = prod.FlexAquire(xnames)
	if (err != nil) {
		t.Errorf("FlexAquire() failed: %v",err)
	}
	srel,err = prod.FlexRelease(xnames)
	if (err != nil) {
		t.Errorf("FlexRelease() failed: %v",err)
	}
	if (srel.Counts.Total != len(xnames)) {
		t.Errorf("FlexRelease() bad total count, exp: %d, got: %d",
			len(xnames),srel.Counts.Total)
	}
	if (srel.Counts.Success != len(xnames)) {
		t.Errorf("FlexRelease() bad success count, exp: %d, got: %d",
			len(xnames),srel.Counts.Success)
	}
	if (srel.Counts.Failure != 0) {
		t.Errorf("FlexRelease() bad failure count, exp: %d, got: %d",
			0,srel.Counts.Success)
	}

	//Test aquire() but force a failure in the server

	failAquire = true
	rsv,err = prod.FlexAquire(xnames)
	if (err != nil) {
		t.Errorf("FlexAquire() failed, should not.")
	}
	failAquire = false

	if ((len(rsv.Success) != (len(xnames) - 1))) {
		t.Errorf("FlexAquire() wrong number of successes: exp: %d, got: %d",
			len(xnames)-1,len(rsv.Success))
	}
	if (len(rsv.Failure) != 1) {
		t.Errorf("FlexAquire() wrong number of failures: exp: %d, got: %d",
			1,len(rsv.Failure))
	}
}

