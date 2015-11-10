//
//
// The Client API runs an HTTPS-server for serving browser clients.
// Upon requests it make internal connections on localhost to
// the MTAContainer and the Backend as required.
//
// A special path prefix go.api/** is for ajax queries by the client
// side Dart application.
//
// Author: Rasmus Winther Zakarias
//

package clientapi


import (
	"net/http"
	"io/ioutil"
	"os"
	"strings"
	"mail.bitlab.dk/mtacontainer"
	"strconv"
	"log"
	"errors"
	"mail.bitlab.dk/utilities"
	"encoding/base64"
	"mail.bitlab.dk/utilities/go"
)

// ---------------------------------------------------------
//
//
// Client API implementation
//
// ---------------------------------------------------------
type ClientAPI struct {
	docRoot         string
	mtacontainer.HealthService;
	events          chan mtacontainer.Event;
	port            int;
	log             *log.Logger;
	validSessions   map[string]string; // sessionId -> username
	pendingMessages []string;
	versionStr	    string;
}


// --------------------------------------------------------
//
// Creates a Server for the website ClientApi
//
//
//
// --------------------------------------------------------
func New(docRoot string, port int) *ClientAPI {
	var result = new(ClientAPI);
	result.docRoot = docRoot;
	result.events = make(chan mtacontainer.Event);
	result.port = port;
	result.log = utilities.GetLogger("[client api] ", os.Stdout);
	result.validSessions = make(map[string]string);
	go result.serve();
	return result;
}

//
// Allow clients to observe events
//
func (a *ClientAPI) GetEvent() chan mtacontainer.Event {
	return a.events;
}

// ---------------------------------------------------------------------
// The HTTPS end-point is configured here
//
// The client app can do five things on the go.api prefix
//
// All other requests are handled by the {viewHandler} serving
// from the file system folder e-mail-service/dartworkspace/build/web.
//
// ---------------------------------------------------------------------

func (a *ClientAPI) serve() {

	versionStr, versionStrErr := ioutil.ReadFile(a.docRoot + "/version.txt");
	if versionStrErr != nil {
		a.versionStr = "No Version";
		a.log.Println(versionStrErr.Error());
	} else {
		a.versionStr = string(versionStr);
	}

	var mux = http.NewServeMux();
	mux.HandleFunc("/go.api/alive", a.alivePingHandler);
	mux.HandleFunc("/go.api/login", a.handleLogin);
	mux.HandleFunc("/go.api/logout", a.logoutHandler);
	mux.HandleFunc("/go.api/sendmail", a.sendMailHandler);
	mux.HandleFunc("/mta/usermessage", a.receiveUserMessage);
	mux.HandleFunc("/", a.viewHandler);


	var addr = ":" + strconv.Itoa(a.port);
	a.events <- mtacontainer.NewEvent(mtacontainer.EK_OK, errors.New("Serving on port: " + addr));

	err := http.ListenAndServeTLS(addr, "cert.pem", "key.pem", mux);

	if (err != nil) {
		log.Fatalln("[ClientApi, Error] " + err.Error());
	}
}

func (a *ClientAPI) receiveUserMessage(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close();

	data, dataErr := ioutil.ReadAll(r.Body);
	if dataErr != nil {
		a.log.Println(" Failed to read data from MTA Server posting a message");
		return;
	}

	a.pendingMessages = append(a.pendingMessages, string(data));
}

//
// Heart-beat allow the client to know ClientAPI is still up.
//
func (ths *ClientAPI) alivePingHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close();

	if len(ths.pendingMessages) > 0 {
		m := ths.pendingMessages[0];
		ths.pendingMessages = ths.pendingMessages[1:];
		w.Write([]byte("!"+m));
	} else {
		w.Write([]byte(ths.versionStr));
	}
}

//
// helper to {handleLogin}
//
func decodeBasicAuth(auth string, log *log.Logger) (string, string, bool) {
// TODO(rwz): We need to handle decoding error
var parts = strings.Split(auth, " ");

if strings.Trim(parts[0], " ") != "Basic" {
log.Println("Not basic: " + parts[0] + ", " + parts[1]);
return "", "", false;
}
var bytes, _ = base64.StdEncoding.DecodeString(strings.Trim(parts[1], " "));
var authorizationString = string(bytes);


var usernameAndPassword = strings.Split(authorizationString, ":");
var username = usernameAndPassword[0];
var password = usernameAndPassword[1];
return username, password, true;
}

//
//
// Forward a login check to the backend
//
func (ths *ClientAPI) handleLogin(w http.ResponseWriter, r *http.Request) {
defer r.Body.Close();

ths.log.Println("Handle Login");

if len(r.Header["Authorization"]) < 1 {
ths.log.Println("Authorization header is missing");
http.Error(w, "Missing authorization", http.StatusBadRequest);
return;
}
authStr := r.Header["Authorization"][0];

var username, password, ok = decodeBasicAuth(authStr, ths.log);
if ok == false {
ths.log.Println("Authorization failed it is not basic ");
http.Error(w, "Bad authorization", http.StatusBadRequest);
return;
}

q := "/login?username=" + username;
q += "&password=" + password;


backendLoginQuery := "http://localhost" + utilities.RECEIVE_BACKEND_LISTEN_FOR_CLIENT_API + q;

resp, err := http.Get(backendLoginQuery);
if err == nil {
ths.log.Println("Ok connection to receiver backend");
sessionId, errAll := ioutil.ReadAll(resp.Body)

if errAll == nil {
if len(sessionId) < 64 {
ths.log.Println("Reporting back to UI that an error occured... " + goh.IntToStr(len(sessionId)));
http.Error(w, resp.Status, resp.StatusCode);
return;
}
ths.log.Println("Adding session id " + string(sessionId));
ths.validSessions[string(sessionId)] = username;
w.Write(sessionId);
return;
}  else {
ths.log.Println("Auch failed to read response body from Receiver backend.");
http.Error(w, "Internal Server Error", http.StatusInternalServerError);
}

} else {
http.Error(w, "Backend refuses login from user", http.StatusInternalServerError);
ths.events <- mtacontainer.NewEvent(mtacontainer.EK_CRITICAL,
errors.New("Could not connect to receiver back end"), ths);
}
}


//
// Handle User logout
//
func (a *ClientAPI) logoutHandler(w http.ResponseWriter, r *http.Request) {
defer r.Body.Close();
var sessionId = r.URL.Query().Get("SessionId");
delete(a.validSessions, sessionId);
}


func (a *ClientAPI) validateSession(sessionId string) bool {

username, ok := a.validSessions[sessionId];
a.log.Println("Sending request from user " + username + " accepted.");
if ok == false {
a.log.Println("Validating sessionId: " + sessionId + " failed");
} else {
a.log.Println("Valid session: " + sessionId);
}
return ok;
}


// ------------------------------------------------
//
// The send mail handler checks that the SesisonId
// in the http-headers is valid. If it is the
// serialized WireMail is forwarded to MTAContainer
// for sending.
//
// ------------------------------------------------
func (a *ClientAPI) sendMailHandler(w http.ResponseWriter, r *http.Request) {

defer r.Body.Close();

a.log.Println("Send Mail Handler - Checking SessionId");

if r.Header.Get("SessionId") == "" {
http.Error(w, "No session id found", http.StatusForbidden);
a.log.Println("Session Id not present");
return;
}

validSession := a.validateSession(r.Header.Get("SessionId"));
if validSession == false {
http.Error(w, "Invalid session", http.StatusForbidden);
a.log.Println("SessionId not valid: " + r.Header.Get("SessionId"));
return;
}

var request, requestError = http.NewRequest("POST", "http://localhost" +
utilities.MTA_LISTENS_FOR_SEND_BACKEND + "/sendmail", r.Body);
if requestError != nil {
http.Error(w, requestError.Error(), http.StatusInternalServerError);
return;
}

if requestError != nil {
http.Error(w, "Failed to create HttpRequest for MTA Container", http.StatusInternalServerError);
a.log.Println("Failed to create HttpRequest for MTA Container:\n" + requestError.Error());
return;
}

response, responseErr := http.DefaultClient.Do(request);
if responseErr != nil {
if response != nil {
http.Error(w, responseErr.Error(), response.StatusCode);
} else {
http.Error(w, responseErr.Error(), http.StatusInternalServerError);
}
a.log.Println(responseErr.Error())
return;
}

if response.StatusCode >= 200 && response.StatusCode < 300 {
a.log.Println("Forwarded e-email to MTA Container with success.");
a.log.Println(response.Status);
} else {
http.Error(w, response.Status, response.StatusCode);
}
}


func loadFile(filename string) ([]byte, error) {
body, err := ioutil.ReadFile(filename)
if err != nil {
return nil, err
}
return body, nil;
}

//
// Handler for serving files from the file system
//
func (s *ClientAPI) viewHandler(w http.ResponseWriter, r *http.Request) {
var extMap map[string]string = map[string]string{"html": "text/html", "css": "text/css", "dart": "application/dart"};

var path = r.URL.Path;
if ("" == path || "/" == path) {
path = "/index.html";
}


filename := s.docRoot + path;
var fi, erro = os.Stat(filename);

if (erro != nil) {
println("[Error] " + erro.Error());
return;
}

if (fi.IsDir() == false) {

println("Serving file: " + filename)
p, e := loadFile(filename)
if (e != nil) {
println("[Error] " + e.Error());
return;
}

var mimeType = "text/plain";
if strings.Contains(filename, ".") {
var ext = filename[strings.LastIndex(filename, ".") + 1:len(filename)];
var v, ok = extMap[ext];
if ok {
mimeType = v;
}
}

w.Header().Add("Content-Type", mimeType);

w.Write(p);

} else {
w.Write([]byte("Directory listings not supported."))
}
}