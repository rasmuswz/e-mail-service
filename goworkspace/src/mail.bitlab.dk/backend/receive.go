//
// The Receive back end listens for the mta container to
// report e-mails ready for delivery and then the ReceiveBackEnd
// stores these in a JsonStore.
//
//
// Author: Rasmus Winter Zakarias
//

package backend
import (
	"mail.bitlab.dk/model"
	"mail.bitlab.dk/mtacontainer"
	"net/http"
	"log"
	"errors"
	"encoding/json"
	"mail.bitlab.dk/utilities"
	"encoding/base64"
	"strings"
	"strconv"
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
	mtacontainer.HealthService;
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
	go res.ListenForMtaContainer();
	go res.StoreMailsInStore();

	return res;
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
	mux.HandleFunc("/getmail", ths.serviceClientApi);
	mux.HandleFunc("/login", ths.handleLogin);
	mux.HandleFunc("/logout", ths.handleLogout);
	mux.HandleFunc("/", ths.handleError);
	err := http.ListenAndServe(utilities.RECEIVE_BACKEND_LISTEN_FOR_CLIENT_API, mux);
	if err != nil {
		log.Fatalln("[Receive BackEnd] Cannot listen for ClientApi:\n" + err.Error());
	}
}
// Client API error request handler
func (ths *ReceiveBackEnd) handleError(w http.ResponseWriter, r * http.Request) {
	defer r.Body.Close();
	ths.log.Println("Error request arrived...");
	http.Error(w,"No such service",http.StatusBadGateway);
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
func (ths *ReceiveBackEnd) createSession(username, location string) string {
	var now = time.Now();
	var sessionId = utilities.HashStringToHex(now.String() + username + location);
	ths.log.Println("Created session id for user: "+username+"\n"+sessionId);
	ths.store.PutJSonBlob(map[string]string{"sessionId": sessionId, "username": username, "location": location});
	return sessionId;
}

// Client API login handler
func (ths *ReceiveBackEnd) handleLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close();
	ths.log.Println("Login Handler on Backend !");
	username := r.URL.Query().Get("username");
	password := r.URL.Query().Get("password");
	location := r.URL.Query().Get("location");

	if (username == "" || password == "" || location == "") {
		http.Error(w,"username, password or location is not set", http.StatusBadRequest);
		return;
	}

	ths.log.Println("Welcome to user " + username + " at " + location);

	users :=  ths.store.GetJSonBlobs(UserBlobNew(username,password).ToJSonMap());

	ths.log.Println("We found " + goh.IntToStr(len(users)) + " entries with this user in our database");
	//
	// If user does not exists at all lets create him :-) That is no blob with "username" set to
	// the given username.
	//
	if (len(users) == 0) {
		ths.log.Println("No user with that password and username, is the any user called " + username + "?")
		users = ths.store.GetJSonBlobs(UserBlobNew(username,password).ToJSonMap());
		var sessionId = ths.createSession(username, location);
		userBlob := UserBlobNewFull(username,password,location,sessionId);
		if len(users) == 0 {
			ths.log.Println("No, lets create " + username);
			ths.store.PutJSonBlob(userBlob.ToJSonMap());
			w.Write([]byte(sessionId));
			ths.log.Println("Just sent "+goh.IntToStr(len(sessionId))+" bytes across");
			return; // success
		} else {
			http.Error(w,"Access Denied",http.StatusForbidden);
		}
		return;
	}

	//
	// Ok we found a user
	//
	if (len(users) == 1) {
		var sessionId = ths.createSession(username, location);
		w.Write([]byte(sessionId));
		ths.store.UpdJSonBlob(UserBlobNew(username,password).ToJSonMap(),UserBlobNewFull(username,password,location,sessionId).ToJSonMap());
	} else {
		http.Error(w,"Access Denied",http.StatusForbidden);

	}
	r.Body.Close();
}

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

func (ths *ReceiveBackEnd) serviceClientApi(w http.ResponseWriter, r *http.Request) {
	type GetMailRequest struct {
		index  int;
		length int;
	}

	var username, ok = CheckAuthorizedUser(ths.store, r);
	if ok == false {
		w.Header()["StatusCode"] = []string{strconv.FormatUint(403, 10)};
		return;
	}

	var jDec = json.NewDecoder(r.Body);
	var ask GetMailRequest = GetMailRequest{};

	emailErr := jDec.Decode(&ask);
	if emailErr != nil {
		log.Println("[Receiver BackEnd] Didn't understand e-mail from ClientApi.");
		return;
	}

	var query = make(map[string]string);
	query["username"] = username;

	var jEnc = json.NewEncoder(w);
	var emailsForUser []map[string]string = ths.store.GetJSonBlobs(query);

	for i := range emailsForUser {
		if (i >= ask.index && i < ask.index + ask.length) {
			var e model.EmailFromJSon = model.EmailFromJSon{};
			e.Content = emailsForUser[i]["content"];
			e.Headers = make(map[string]string)
			for k, v := range emailsForUser[i] {
				if k != "content" {
					e.Headers[k] = v;
				}
			}
			err := jEnc.Encode(&e);
			if err != nil {
				log.Println("Failed to encode and send e-mail to ClientApi.");
			}
		}
	}

	log.Println("[Receive Backend] Incoming request from the client API, leaving");
}

func (ths *ReceiveBackEnd) ListenForMtaContainer() {

	var mux = http.NewServeMux();
	mux.HandleFunc("/newmail", ths.receiveMail);
	var err = http.ListenAndServe("localhost" + utilities.RECEIVE_BACKEND_LISTENS_FOR_MTA, mux);
	if err != nil {
		log.Fatalln("[Receiver Backend] Failed to listen for MTA: " + err.Error());
	}
	log.Println("[Receiver Backend] Leaving we no longer listens for incoming mail.");
}

func (ths *ReceiveBackEnd) receiveMail(w http.ResponseWriter, r *http.Request) {
	ths.events <- mtacontainer.NewEvent(mtacontainer.EK_OK, errors.New("Incoming e-mail from Mta"));
	var jDecoder = json.NewDecoder(r.Body);
	var jemail model.EmailFromJSon;
	var jemailErr = jDecoder.Decode(&jemail);

	if (jemailErr != nil) {
		ths.events <- mtacontainer.NewEvent(mtacontainer.EK_DOWN_TEMPORARILY, jemailErr);
	} else {
		var email = model.NewEmailFromJSon(&jemail);
		ths.incoming <- email;
	}
}

func (ths *ReceiveBackEnd) StoreMailsInStore() {
	for {
		select {
		case mail := <-ths.incoming:
			var mailHeaders = mail.GetHeaders();
			var users = ths.store.GetJSonBlobs(map[string]string{"email": mailHeaders[model.EML_HDR_TO]});
			for i := range users {
				var blob = make(map[string]string);
				blob["content"] = mail.GetContent();
				blob["mbox"] = model.MBOX_NAME_INBOX;
				blob["username"] = users[i]["username"];
				for k, v := range mail.GetHeaders() {
					blob[k] = v;
				}
				ths.store.PutJSonBlob(blob);
			}
		case cmd := <-ths.cmd:
			if (cmd == CMD_SHUTDOWN) {
				return;
			}
		}
	}
}