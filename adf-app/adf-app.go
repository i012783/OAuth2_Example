package main


import(
	//"log"
	"net/http"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/net/context"
	"os"
	"io/ioutil"
	"encoding/base64"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////////////////

const (
	DEBUG = 0
	htmlIndex = "<html><body><center><div style=\"font-size:50;color:red\">ADF Application</div><p><p><form method=\"get\" action=\"/login\"><button style=\"font-size:20px;height:50px;width:200px\" type=\"submit\">Get ADF Data</button></form></center></body></html>"
	CLIENT_ID = "adf-app"
	CLIENT_SECRET = "password"
	CHECK_TOKEN = "https://a5b8c64c-f9d2-4eef-8a78-53ca5e77ee44.predix-uaa.run.asv-pr.ice.predix.io/check_token"
	AUTH_URL = "https://a5b8c64c-f9d2-4eef-8a78-53ca5e77ee44.predix-uaa.run.asv-pr.ice.predix.io/oauth/authorize"
	TOKEN_URL = "https://a5b8c64c-f9d2-4eef-8a78-53ca5e77ee44.predix-uaa.run.asv-pr.ice.predix.io/oauth/token"
	ADF1 ="https://adf1.run.asv-pr.ice.predix.io"
	REDIRECT_URL = "https://adf-app.run.asv-pr.ice.predix.io/authcode"
)

/*   LOCAL ENDPOINTS   */
/*const (
	CHECK_TOKEN = "http://localhost:8080/uaa/check_token"
	AUTH_URL = "http://localhost:8080/uaa/oauth/authorize"
	TOKEN_URL = "http://localhost:8080/uaa/oauth/token"
	ADF1 = "http://localhost:3000"
	REDIRECT_URL = "http://localhost:8082/authcode"
)*/


/*   OAUTH2 SCOPES   */
var (
	SCOPES = []string{"behr"}
	OAUTH_STATE_STRING = "123456"
)

/*   UAA OAUTH2 ENDPOINTS   */
var  ENDPOINT = oauth2.Endpoint{
	AuthURL:  AUTH_URL,
	TokenURL: TOKEN_URL,
}

/*   Create OAUTH2 configuration client to use in Access Code/Token generation   */
var (
	oauthConf = &oauth2.Config{

		ClientID: 	CLIENT_ID,
		ClientSecret: 	CLIENT_SECRET,
		Scopes:       	SCOPES,
		Endpoint:      	ENDPOINT,
		RedirectURL:    REDIRECT_URL,
	}

	/*   Random GUID used to check for CSRF   */
	oauthStateString = OAUTH_STATE_STRING

)
////////////////////////////////////////////////////////////////////////////////////////////

func Login(w http.ResponseWriter, r *http.Request){

	// AUTHORIZATION_CODE GRANT
	url := oauthConf.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusFound)
}


func CallBack(w http.ResponseWriter, r *http.Request) {


	// AUTHORIZATION_CODE GRANT
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if state != oauthStateString {
		fmt.Fprintf(w, "invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	token, _ := oauthConf.Exchange(context.TODO(), code)

	client := oauthConf.Client(context.Background(), token)
	req, _ := http.NewRequest("GET", ADF1 + "/engines", nil)
	resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}

	// PRINT USER JWT
	fmt.Fprintf(w, "<div style=\"font-size:25;color:black;\">--------------------ADF-APP--------------------</div>")
	fmt.Fprintf(w, "<div style=\"font-size:20;color:black;\">1.) Generate User JWT Token - \"authorization_code\" Grant Type</div><br>")
	fmt.Fprintf(w, "<div style=\"font-size:20;color:black\">User JSON Web Token (JWT)</div><br>" )
	fmt.Fprintf(w, "<div style=\"width:1200px;word-wrap:break-word;\" NAME=\"SOFT\" WRAP=HARD><div style=\"font-size=20;\" color=\"black\">" + token.AccessToken + "</div></div><br>")

	// DECODE TOKEN
	result := strings.Split(token.AccessToken, ".")
	d, _ := base64.StdEncoding.DecodeString(result[1])
	fmt.Fprintf(w, "<div style=\"font-size:20;color:black\">Decoded User JWT</div><br>" )
	fmt.Fprintf(w, "<div style=\"width:1200px;word-wrap:break-word;font-size:20;color:black\">" + string(d) + "</div></div><br>")
	w.Write([]byte(contents))
}

func Index(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, htmlIndex)
}


func main() {

	mux := mux.NewRouter()
	mux.HandleFunc("/", Index)
	mux.HandleFunc("/login", Login)
	mux.HandleFunc("/authcode", CallBack)
	http.Handle("/", mux)

	http.ListenAndServe(":" + os.Getenv("PORT"), nil)
	//log.Println("Listening.......")
	//http.ListenAndServe(":8082", nil)

}