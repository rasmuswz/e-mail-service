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
)

type ClientAPI struct {
	docRoot string
	mtacontainer.HealthService;
	events  chan mtacontainer.Event;
	port    int;
}


func (a *ClientAPI) GetEvent() chan mtacontainer.Event {
	return a.events;
}


// --------------------------------------------------------
//
// Create a Server for the website and ClientApi
//
// --------------------------------------------------------
func NewServer(docRoot string, port int) *ClientAPI {
	var result = new(ClientAPI);
	result.docRoot = docRoot;
	result.events = make(chan mtacontainer.Event);
	result.port = port;
	go result.serve();
	return result;
}

func (ths *ClientAPI) alivePingHandler(w http.ResponseWriter, r *http.Request) {
	version := utilities.GetString("https://localhost/version.txt");
	w.Write([]byte(version));
	w.Header()["StatusCode"] = []string{"200"};
	r.Body.Close();
}

func decodeBasicAuth(auth string) (string,string,bool) {
	// TODO(rwz): We need to handle decoding error
	var bytes,_ = base64.StdEncoding.DecodeString(auth);
	var authstr = string(bytes);
	var parts = strings.Split(authstr," ");
	if strings.Trim(parts[0]," ") != "Basic" {
		return "","",false;
	}
	var usernameAndPassword = strings.Split(parts[1],":");
	var username = usernameAndPassword[0];
	var password = usernameAndPassword[1];

	return username,password,true;
}

func (ths *ClientAPI) handleLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close();

	if len(r.Header["Authorization"]) < 1 {
		http.Error(w,"missing authorization",http.StatusBadRequest);
		r.Body.Close();
		return;
	}
	authStr := r.Header["Authorization"][0];

	var username,password,ok = decodeBasicAuth(authStr);
	if ok == false {
		http.Error(w,"Bad authorization",http.StatusBadRequest);
		return;
	}

	q := "?username"+username;
	q += "&password"+password;
	q += "&location="+r.URL.Query().Get("location");

	resp, err := http.Get("http://localhost" + utilities.RECEIVE_BACKEND_LISTEN_FOR_CLIENT_API + q);
	if err != nil {
		sessionId, errAll := ioutil.ReadAll(resp.Body)
		if errAll == nil {
			w.Write(sessionId);
			r.Body.Close();
			return;
		}
		http.Error(w,"Internal server error decoding response from Receiver Back End.",500);
	}
	http.Error(w,"Internal server error Receiver Back End is down",500);
	ths.events <- mtacontainer.NewEvent(mtacontainer.EK_CRITICAL,
		errors.New("Could not connect to receiver back end"),ths);
}

func (a *ClientAPI) logoutHandler(w http.ResponseWriter, r *http.Request) {
	var q = "/logout?session="+r.URL.Query().Get("session");
	_,err := http.Get("http://localhost" + utilities.RECEIVE_BACKEND_LISTEN_FOR_CLIENT_API + q);

	if err != nil {
		w.Write([]byte("OK"));
	}
	r.Body.Close();
}

func (a *ClientAPI) getMboxesHandler(w http.ResponseWriter, r * http.Request) {

}

func (a *ClientAPI) serve() {

	var mux = http.NewServeMux();
	mux.HandleFunc("/go.api/alive/", a.alivePingHandler);
	mux.HandleFunc("/go.api/login",  a.handleLogin);
	mux.HandleFunc("/go.api/logout", a.logoutHandler);
	mux.HandleFunc("/go.api/mboxes", a.getMboxesHandler);

	mux.HandleFunc("/", a.viewHandler);


	var addr = ":" + strconv.Itoa(a.port);
	a.events <- mtacontainer.NewEvent(mtacontainer.EK_OK, errors.New("Serving on port: " + addr));

	err := http.ListenAndServeTLS(addr, "cert.pem", "key.pem", mux);

	if (err != nil) {
		log.Fatalln("[ClientApi, Error] " + err.Error());
	}
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