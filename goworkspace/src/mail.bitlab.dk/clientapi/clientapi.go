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
	"encoding/hex"
	"bytes"
)

// ---------------------------------------------------------
//
//
// Client API implementation
//
// ---------------------------------------------------------
type ClientAPI struct {
	docRoot string
	mtacontainer.HealthService;
	events  chan mtacontainer.Event;
	port    int;
	log *log.Logger;
}


// --------------------------------------------------------
//
// Create a Server for the website and ClientApi
//
//
//
// --------------------------------------------------------
func New(docRoot string, port int) *ClientAPI {
	var result = new(ClientAPI);
	result.docRoot = docRoot;
	result.events = make(chan mtacontainer.Event);
	result.port = port;
	result.log = utilities.GetLogger("[client api] ",os.Stdout);
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

	var mux = http.NewServeMux();
	mux.HandleFunc("/go.api/alive", a.alivePingHandler);
	mux.HandleFunc("/go.api/login",  a.handleLogin);
	mux.HandleFunc("/go.api/logout", a.logoutHandler);
	mux.HandleFunc("/go.api/sendmail", a.sendMailHandler);
	mux.HandleFunc("/go.api/getmail", a.getMailHandler);

	mux.HandleFunc("/", a.viewHandler);


	var addr = ":" + strconv.Itoa(a.port);
	a.events <- mtacontainer.NewEvent(mtacontainer.EK_OK, errors.New("Serving on port: " + addr));

	err := http.ListenAndServeTLS(addr, "cert.pem", "key.pem", mux);

	if (err != nil) {
		log.Fatalln("[ClientApi, Error] " + err.Error());
	}
}


//
// Heart-beat allow the client to know ClientAPI is still up.
//
func (ths *ClientAPI) alivePingHandler(w http.ResponseWriter, r *http.Request) {
	version := utilities.GetString("https://localhost/version.txt");
	w.Write([]byte(version));
	w.Header()["StatusCode"] = []string{"200"};
	r.Body.Close();
}

//
// helper to {handleLogin}
//
func decodeBasicAuth(auth string, log *log.Logger) (string,string,bool) {
	// TODO(rwz): We need to handle decoding error
	var parts = strings.Split(auth," ");

	if strings.Trim(parts[0]," ") != "Basic" {
		log.Println("Not basic: "+parts[0]+", "+parts[1]);
		return "","",false;
	}
	var bytes,_ = base64.StdEncoding.DecodeString(strings.Trim(parts[1]," "));
	var authorizationString = string(bytes);


	var usernameAndPassword = strings.Split(authorizationString,":");
	var username = usernameAndPassword[0];
	var password = usernameAndPassword[1];
	return username,password,true;
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
		http.Error(w,"Missing authorization",http.StatusBadRequest);
		return;
	}
	authStr := r.Header["Authorization"][0];

	var username,password,ok = decodeBasicAuth(authStr,ths.log);
	if ok == false {
		ths.log.Println("Authorization failed it is not basic ");
		http.Error(w,"Bad authorization",http.StatusBadRequest);
		return;
	}

	q := "/login?username="+username;
	q += "&password="+password;


	receiveBackendLoginQuery := "http://localhost" + utilities.RECEIVE_BACKEND_LISTEN_FOR_CLIENT_API + q;

	resp, err := http.Get(receiveBackendLoginQuery);
	if err == nil {
		ths.log.Println("Ok connection to receiver backend");
		sessionId, errAll := ioutil.ReadAll(resp.Body)
		if len(sessionId) < 64 {
			ths.log.Println("Reporting back to UI that an error occured... "+goh.IntToStr(len(sessionId)));
			http.Error(w,resp.Status,resp.StatusCode);
			return;
		} else {
			ths.log.Println("We have sessionId with length: "+goh.IntToStr(len(sessionId)));
			ths.log.Println("\n"+hex.Dump(sessionId));
		}


		if errAll == nil {
			w.Write(sessionId);
			return;
		}  else {
			ths.log.Println("Auch failed to read response body from Receiver backend.");
			http.Error(w,"Internal Server Error",http.StatusInternalServerError);
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
	var q = "/logout?session="+r.URL.Query().Get("session");
	_,err := http.Get("http://localhost" + utilities.RECEIVE_BACKEND_LISTEN_FOR_CLIENT_API + q);

	if err != nil {
		w.Write([]byte("OK"));
	}
	r.Body.Close();
}


func (a *ClientAPI) sendMailHandler(w http.ResponseWriter, r * http.Request) {

	defer r.Body.Close();

	a.log.Println("Sending mail");

	rawdata,dataErr := ioutil.ReadAll(r.Body);
	if dataErr != nil {
		http.Error(w,"Could not read body of request.",http.StatusBadRequest);
		return;
	}

	a.log.Println(goh.IntToStr(len(rawdata)));
	a.log.Println("Forwarding data to MTA container:\n"+hex.Dump(rawdata));


	var request, requestError = http.NewRequest("POST", "http://localhost"+
		utilities.MTA_LISTENS_FOR_SEND_BACKEND+"/sendmail",bytes.NewBuffer(rawdata));
	if requestError != nil {
		http.Error(w,"Could not create request",http.StatusInternalServerError);
		return;
	}

	request.Header["Content-Length"] = []string{goh.IntToStr(len(rawdata))};


	if requestError != nil {
		http.Error(w, "Failed to create HttpRequest for MTA Container", http.StatusInternalServerError);
		a.log.Println("Failed to create HttpRequest for MTA Container:\n"+requestError.Error());
		return;
	}

	response, responseErr := http.DefaultClient.Do(request);
	if responseErr != nil {
		if response != nil {
			http.Error(w, responseErr.Error(), response.StatusCode);
		} else {
			http.Error(w,responseErr.Error(), http.StatusInternalServerError);
		}
		a.log.Println(responseErr.Error())
		return;
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		a.log.Println("Forwarded e-email to MTA Container with success.");
		a.log.Println(response.Status);
	}
}


// ---------------------------------------------------------
//
// Get Emails for logged in user. The query from the client
// is forwarded to the Back-end for looking up INBOX entries
// in the JSon-storage.
//
// ---------------------------------------------------------

func (a *ClientAPI)getMailHandler(w http.ResponseWriter, r * http.Request) {

	defer r.Body.Close();

	if len(r.Header["SessionId"]) < 1 {
		a.log.Println("No access with out an session id")
		http.Error(w,"Session ID missing access denied.",http.StatusForbidden);
		return;
	}
	ses := r.Header["sessionId"][0];

	req, reqErr := http.NewRequest("GET",utilities.RECEIVE_BACKEND_LISTEN_FOR_CLIENT_API+"/getmail?SessionId="+ses,r.Body);
	if reqErr != nil {
		http.Error(w,"Could not create request: "+reqErr.Error(),http.StatusInternalServerError);
		return;
	}

	resp,err := http.DefaultClient.Do(req);
	if err != nil {
		http.Error(w,err.Error(),http.StatusInternalServerError);
		return;
	}

	data, dataReq := ioutil.ReadAll(resp.Body);
	if dataReq != nil {
		http.Error(w,"Damn it could not get response from Receiver Back-end",http.StatusInternalServerError);
		return;
	}

	log.Println("We have successfully forwardet a list of emails.");
	w.Write(data);
	return;

}


func (a *ClientAPI) clientApiHandler(w http.ResponseWriter, r *http.Request) {

	var path string = r.URL.Path;

	if path[strings.LastIndex(path, "/") + 1:len(path)] == "alive" {
		w.Write([]byte("yes"));
		r.Body.Close();
	}

}

func loadFile(filename string) ([]byte, error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return body, nil;
}

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