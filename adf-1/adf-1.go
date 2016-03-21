/*
POC: 212327125
Date:  2/28/16
 */

package main

import(
	"fmt"
	"net/http"
	"os"
	//"log"
	"github.com/gorilla/mux"
	"io/ioutil"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/net/context"
	"github.com/codegangsta/negroni"
	"time"
	"net/url"
	"bytes"
	"golang.org/x/oauth2"
	"encoding/base64"
	"strings"
	"security.aviation.ge.com/core"
)

//////////////////////////////////////////

const(
	DEBUG = 0
	SLEEP = 1
	CLIENT_ID = "adf-1"
	CLIENT_SECRET = "password"
	SERVICE_CREDENTIAL = "YWRmLTE6cGFzc3dvcmQ="
	ACS_EP = "http://acs.somewhere.com"
	CHECK_TOKEN = "https://a5b8c64c-f9d2-4eef-8a78-53ca5e77ee44.predix-uaa.run.asv-pr.ice.predix.io/check_token"
	TOKEN_URL = "https://a5b8c64c-f9d2-4eef-8a78-53ca5e77ee44.predix-uaa.run.asv-pr.ice.predix.io/oauth/token"
)

/*

const (
	CHECK_TOKEN = "http://localhost:8080/uaa/check_token"
	TOKEN_URL = "http://localhost:8080/uaa/oauth/token"
	ADF2_EP = "http://localhost:3001/sn"
)
*/

//////////////  Scopes
var (
	SCOPES = []string{""}
	ADF2_EP = "https://adf2.run.asv-pr.ice.predix.io/sn"
)

//////////////////////////////////////////


func IsAuthenticated() negroni.Handler  {
	au := func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

		if len(r.Header.Get("Authorization")) == 0  {
			fmt.Fprintf(w, "<div style=\"font-size:25;color:black;\">No Direct Access - Please Login</div>")
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


		if DEBUG != 0 {
			time.Sleep(SLEEP * time.Second)
		}

		resp, err := client.Do(req)

		if resp.StatusCode != 200{
			//Handle the different response codes appropriately
			fmt.Fprintf(w, "<div style=\"font-size:20;color:red\">USER Token Invalid or Expired</div><br>")

			return
		}else{
			fmt.Fprintf(w, "<div style=\"font-size:25;color:black;\">--------------------ADF-1 IsAuthenticated()--------------------</div>")
			fmt.Fprintf(w, "<div style=\"font-size:20;color:black;\">2.) User JWT Valid</div>")
		}
		next(w,r)
	}
	return negroni.HandlerFunc(au)
}

func IsAuthorized(AccessQuery string) negroni.Handler  {
	az := func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		fmt.Fprintf(w, "<div style=\"font-size:25;color:black\">--------------------ADF-1 IsAuthorized()-----------------------</div>")
		fmt.Fprintf(w, "<div style=\"font-size:20;color:red;\">3. Build ACS entitlement query</div>")
		fmt.Fprintf(w, "<div style=\"font-size:20;color:red\">4. Create REST client and call ACS</div>" )
		fmt.Fprintf(w, "<div style=\"font-size:20;color:red\">5. PERMIT/DENY whether user is allowed to use ADF-1 API</div><br>" )
		next(w,r)
	}
	return negroni.HandlerFunc(az)
}


func Engines(w http.ResponseWriter, r *http.Request){

	fmt.Fprintf(w, "<div style=\"font-size:25;color:black\">----------------------ADF-1 \"/engines\"-------------------------</div>")
	fmt.Fprintf(w, "<div style=\"font-size:20;color:black;\">6. Generate ADF-1 API Token - \"client_credential\" Grant Type</div><br>")


	config := clientcredentials.Config{
		ClientID: CLIENT_ID,
		ClientSecret: CLIENT_SECRET,
		TokenURL: TOKEN_URL,
		Scopes: SCOPES,
	}

	// GET TOKEN TO DECODE
	t := &oauth2.Token{}
	t, _ = config.Token(context.Background())
	fmt.Fprintf(w, "<div style=\"font-size:20;color:black\">ADF-1 API Token</div><br>" )
	fmt.Fprintf(w, "<div style=\"width:1200px;word-wrap:break-word;\" NAME=\"SOFT\" WRAP=HARD><div style=\"font-size=20;\" color=\"black\">" + t.AccessToken + "</div></div><br><br>")

	result := strings.Split(t.AccessToken, ".")
	d, _ := base64.StdEncoding.DecodeString(result[1])
	fmt.Fprintf(w, "<div style=\"font-size:20;color:black\">Decoded ADF-1 API JWT</div><br>" )
	fmt.Fprintf(w, "<div style=\"width:1200px;word-wrap:break-word;font-size:20;color:black\">" + string(d) + "</div></div><br>")
	//

	// GENERATE NEW REST CLIENT TO GET DATA - the client will update its token if it's expired
	client := config.Client(context.Background())
	resp, err := client.Get(ADF2_EP)
		if err != nil{
			panic(err)
		}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	  if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	  }else{
		w.Write([]byte(contents))
	  	}
}

func ParseTokenRequest(r *http.Request) string{
	token := r.Header.Get("Authorization")[7:]
	return token
}

func Test(w http.ResponseWriter, r *http.Request){

	if len(r.Header.Get("Authorization")) == 0  {
		fmt.Fprintf(w, "Authorization Header Missing - No direct access")
		return
	}

	token := ParseTokenRequest(r)
	fmt.Fprintf(w, "ADF1: " + token)
}

func main() {

	mux := mux.NewRouter()
	mux.PathPrefix("/test").Methods("GET").Handler(negroni.New(negroni.Handler(IsAuthenticated()), negroni.Handler(core.IsAuthorized("http://localhost:3001/sn")), negroni.Wrap(http.HandlerFunc(Test))))
	mux.PathPrefix("/engines").Methods("GET").Handler(negroni.New(negroni.Handler(IsAuthenticated()), negroni.Handler(core.IsAuthorized("http://localhost:3001/sn")), negroni.Wrap(http.HandlerFunc(Engines))))

	http.Handle("/", mux)
	//log.Println("Listening.....")
	http.ListenAndServe(":" + os.Getenv("PORT"), nil)
	//http.ListenAndServe(":3000", nil)



}
