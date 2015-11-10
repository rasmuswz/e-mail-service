//
// The Back-End services the Client-API assisting with storing user names
// and log in related informatin.
//
// Author: Rasmus Winter Zakarias
//

package backend
import (
	"mail.bitlab.dk/model"
	"mail.bitlab.dk/mtacontainer"
	"net/http"
	"log"
	"mail.bitlab.dk/utilities"
	"encoding/base64"
	"strings"
	"time"
	"os"
	"mail.bitlab.dk/utilities/go"
)

const (
	CMD_SHUTDOWN = 0x01
)

// ---------------------------------------------------------------
//
// Receive Back-End implementation
//
// ---------------------------------------------------------------
type ReceiveBackEnd struct {
	store    JSonStore;
	incoming chan model.Email;
	events   chan mtacontainer.Event;
	cmd      chan int;
	log      *log.Logger;
}

func NewReceiveBackend(store JSonStore) *ReceiveBackEnd {
	var res *ReceiveBackEnd = new(ReceiveBackEnd);
	res.store = store;
	res.incoming = make(chan model.Email);
	res.events = make(chan mtacontainer.Event);
	res.cmd = make(chan int);
	res.log = utilities.GetLogger("[Backend Server]", os.Stdout);
	go res.ListenForClientApi();

	return res;
}


func (ths *ReceiveBackEnd) GetEvent() chan mtacontainer.Event {
	return ths.events;
}


func (ths *ReceiveBackEnd) Stop() {
	ths.cmd <- CMD_SHUTDOWN;
}


// ---------------------------------------------------------
//
// Client API Functionality servicing entry point
// http://localhost:10301
//
// ---------------------------------------------------------
func (ths *ReceiveBackEnd) ListenForClientApi() {
	var mux = http.NewServeMux();
	mux.HandleFunc("/login", ths.handleLogin);
	mux.HandleFunc("/logout", ths.handleLogout);
	mux.HandleFunc("/", ths.handleError);
	err := http.ListenAndServe(utilities.RECEIVE_BACKEND_LISTEN_FOR_CLIENT_API, mux);
	if err != nil {
		log.Fatalln("[Receive BackEnd] Cannot listen for ClientApi:\n" + err.Error());
	}
}
// Client API error request handler
func (ths *ReceiveBackEnd) handleError(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close();
	ths.log.Println("Error request arrived...");
	http.Error(w, "No such service", http.StatusBadGateway);
}

// Client API logout request handler
func (ths *ReceiveBackEnd) handleLogout(w http.ResponseWriter, r *http.Request) {
	var sessionId = r.URL.Query().Get("session");
	ths.log.Println("Handle logout... ");
	ths.log.Println("Session: " + sessionId + " is terminating.");
	ths.store.GetJSonBlobs(map[string]string{"SessionId": sessionId});
	r.Body.Close();
	return;
}

// Client API helper function to create session identification string
func (ths *ReceiveBackEnd) createSession(username string) string {
	var now = time.Now();
	var sessionId = utilities.HashStringToHex(now.String() + username );
	ths.log.Println("Created session id for user: " + username + "\n" + sessionId);
	ths.store.PutJSonBlob(map[string]string{"sessionId": sessionId, "username": username});
	return sessionId;
}

// Client API login handler
func (ths *ReceiveBackEnd) handleLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close();
	ths.log.Println("Login Handler on Backend !");
	username := r.URL.Query().Get("username");
	password := r.URL.Query().Get("password");


	if (username == "" || password == "") {
		http.Error(w, "username, password or location is not set", http.StatusBadRequest);
		return;
	}

	ths.log.Println("Welcome to user " + username + " at ");

	users := ths.store.GetJSonBlobs(UserBlobNew(username, password).ToJSonMap());

	ths.log.Println("We found " + goh.IntToStr(len(users)) + " entries with this user in our database");
	//
	// If user does not exists at all lets create him :-) That is no blob with "username" set to
	// the given username.
	//
	if (len(users) == 0) {
		ths.log.Println("No user with that password and username, is the any user called " + username + "?")
		users = ths.store.GetJSonBlobs(UserBlobNew(username, "").ToJSonMap());
		var sessionId = ths.createSession(username);
		userBlob := UserBlobNewFull(username, password, sessionId);
		if len(users) == 0 {
			ths.log.Println("No, lets create " + username);
			ths.store.PutJSonBlob(userBlob.ToJSonMap());
			ths.store.PutJSonBlob(NewMBox(username, model.MBOX_NAME_INBOX).ToJSonMap());
			w.Write([]byte(sessionId));
			ths.log.Println("Just sent " + goh.IntToStr(len(sessionId)) + " bytes across");
			return; // success
		} else {
			http.Error(w, "Access Denied", http.StatusForbidden);
		}
		return;
	}

	//
	// Ok we found a user
	//
	if (len(users) == 1) {
		var sessionId = ths.createSession(username);
		w.Write([]byte(sessionId));
		ths.store.UpdJSonBlob(UserBlobNew(username, password).ToJSonMap(), UserBlobNewFull(username, password, sessionId).ToJSonMap());
	} else {
		http.Error(w, "Access Denied", http.StatusForbidden);

	}
	r.Body.Close();
}

//
// Check user against Json-store
//
func CheckAuthorizedUser(store JSonStore, req *http.Request) (string, bool) {
	var credentials = req.Header["Authorization"][0];
	var decoded, decodedErr = base64.StdEncoding.DecodeString(credentials)
	if decodedErr != nil {
		log.Println("error could not decode credentials");
		return "", false;
	}
	var s = string(decoded);
	var parts = strings.Split(s, ":");
	var username = parts[0];
	var password = parts[1];

	var res []map[string]string = store.GetJSonBlobs(map[string]string{"type":"user", "username": username, "password": password});

	if (len(res) == 0) {
		log.Println("User authenticated.");
		return username, true;
	} else {
		log.Println("User access Denied: " + username);
		return username, false;
	}
}

