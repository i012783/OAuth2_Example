/*
POC: 212327125
Date:  2/28/16
 */


package main

import(
	"os"
	//"log"
	"net/http"
	"github.com/gorilla/mux"
	"net/url"
	"bytes"
	"fmt"
	"time"
	"github.com/codegangsta/negroni"
	"strconv"
	"encoding/json"

)

////////////////////////////////////////////////////////////////////////////////////////////

const (
	DEBUG = 0
	SLEEP = 1
	SERVICE_CREDENTIAL = "YWRmLTI6cGFzc3dvcmQ="
	//ACS_EP = "http://acs.somewhere.com"
	//CHECK_TOKEN = "https://41557c01-237d-4bb4-a6dd-0d59e46aa6f1.predix-uaa.run.asv-pr.ice.predix.io/check_token"
	//CHECK_TOKEN = "https://59ae3a32-1d26-4f8f-abd8-02a3be95be0a.predix-uaa.run.asv-pr.ice.predix.io/check_token"
	//CHECK_TOKEN = "http://localhost:8080/uaa/check_token"
	CHECK_TOKEN = "https://a5b8c64c-f9d2-4eef-8a78-53ca5e77ee44.predix-uaa.run.asv-pr.ice.predix.io/check_token"
)


////////////////////////////////////////////////////////////////////////////////////////////

type SerialNumber struct {
	Sn 		string `json:"sn"`
	Tail		string	`json:"tail"`
	Location	string `json:"location"`

}


func IsAuthenticated() negroni.Handler  {
	au := func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {


		if len(r.Header.Get("Authorization")) == 0  {
			fmt.Fprintf(w, "Service can only be called by ADF1 - No direct access")
			return
		}

		token := ParseTokenRequest(r)

		data := url.Values{}
		data.Set("token", token)

		client := &http.Client{}
		req, err := http.NewRequest("POST", CHECK_TOKEN, bytes.NewBufferString(data.Encode()))
		if err != nil {
			panic(err)
		}

		req.Header.Add("Authorization", "Basic " + SERVICE_CREDENTIAL)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")


		if DEBUG !=0 {
			time.Sleep(SLEEP * time.Second)
		}

		resp, err := client.Do(req)
		if resp.StatusCode != 200{
			//Handle the different response codes appropriately
			fmt.Fprintf(w, "Status Code SWK: " + strconv.Itoa(resp.StatusCode))
			fmt.Fprintf(w, "<div style=\"font-size:25;color:red\">ADF-1 API Token Invalid or Expired - STATUS: " + strconv.Itoa(resp.StatusCode) + "</div><p><p>")
			return
		}

		if DEBUG == 1 {
			fmt.Fprintf(w, "Service ADF2:")
			fmt.Fprintf(w, "    Client_Credentials from ADF1 - " + token)
		}else{
			fmt.Fprintf(w, "<div style=\"font-size:25;color:black;\">--------------------ADF-2 IsAuthenticated()--------------------</div>")
			fmt.Fprintf(w, "<div style=\"font-size:20;color:black;\">7.) ADF-1 API JWT Valid</div>")
		}
		next(w,r)
	}
	return negroni.HandlerFunc(au)
}

func ParseTokenRequest(r *http.Request) string{
	token := r.Header.Get("Authorization")[7:]
	return token
}

func IsAuthorized(AccessQuery string) negroni.Handler  {
	az := func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		fmt.Fprintf(w, "<div style=\"font-size:25;color:black\">--------------------ADF-2 IsAuthorized()-----------------------</div>")
		fmt.Fprintf(w, "<div style=\"font-size:20;color:red;\">8. Build ACS entitlement query</div>")
		fmt.Fprintf(w, "<div style=\"font-size:20;color:red\">9. Create REST client and call ACS</div>" )
		fmt.Fprintf(w, "<div style=\"font-size:20;color:red\">10. PERMIT/DENY whether user is allowed to use ADF-2 API</div>" )
		next(w,r)
	}
	return negroni.HandlerFunc(az)
}


func GetSN(w http.ResponseWriter, r *http.Request){

	fmt.Fprintf(w, "<div style=\"font-size:25;color:black\">--------Return data if Authenticated & Authorized--------</div><p><p>")
	sn := SerialNumber{"1234567890", "T223X", "Cincinnati"}
	ka, _:= json.Marshal(sn)
	fmt.Fprintf(w, "<div style=\"font-size:20;color:blue\">" + string(ka) + "</div>")
}


func main(){

	mux := mux.NewRouter()
	mux.PathPrefix("/sn").Methods("GET").Handler(negroni.New(negroni.Handler(IsAuthenticated()), negroni.Handler(IsAuthorized("http://localhost:3001/sn")), negroni.Wrap(http.HandlerFunc(GetSN))))
	http.Handle("/", mux)
	//log.Println("Listening........")

	http.ListenAndServe(":" + os.Getenv("PORT"), nil)
	//http.ListenAndServe(":3001", nil)

}