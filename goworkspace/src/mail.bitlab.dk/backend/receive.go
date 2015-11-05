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
}

func NewReceiveBackend(store JSonStore) *ReceiveBackEnd{
	var res *ReceiveBackEnd = new(ReceiveBackEnd);
	res.store = store;
	res.incoming = make(chan model.Email);
	res.events = make(chan mtacontainer.Event);
	res.cmd = make(chan int);
	go res.ListenForClientApi();
	go res.ListenForMtaContainer();
	go res.StoreMailsInStore();

	return res;
}

func (ths *ReceiveBackEnd) Stop() {
	ths.cmd <- CMD_SHUTDOWN;
}

func (ths *ReceiveBackEnd) ListenForClientApi() {
	var mux = http.NewServeMux();
	mux.HandleFunc("/getmail",ths.serviceClientApi);
	err := http.ListenAndServe(utilities.RECEIVE_BACKEND_LISTEN_FOR_CLIENT_API,mux);
	if err != nil {
		log.Fatalln("[Receive BackEnd] Cannot listen for ClientApi:\n"+err.Error());
	}
}

func CheckAuthorizedUser(store JSonStore,req *http.Request ) (string,bool) {
	var credentials = req.Header["Authorization"][0];
	var decoded, decodedErr = base64.StdEncoding.DecodeString(credentials)
	if decodedErr != nil {
		log.Println("error could not decode credentials");
		return "",false;
	}
	var s = string(decoded);
	var parts = strings.Split(s,":");
	var username = parts[0];
	var password = parts[1];

	var res []map[string]string = store.GetJSonBlob(map[string]string{"type":"user", "username": username, "password": password});

	if (len(res) == 0) {
		log.Println("User authenticated.");
		return username,true;
	} else {
		log.Println("User access Denied: "+username);
		return username,false;
	}
}

func (ths *ReceiveBackEnd) serviceClientApi(w http.ResponseWriter, r *http.Request) {
	type GetMailRequest struct {
		index int;
		length int;
	}

	var username, ok = CheckAuthorizedUser(ths.store,r);
	if  ok == false {
		w.Header()["StatusCode"] = []string{strconv.FormatUint(403,10)};
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
	var emailsForUser []map[string]string = ths.store.GetJSonBlob(query);

	for i := range emailsForUser {
		if (i >= ask.index && i < ask.index + ask.length) {
			var e model.EmailFromJSon = model.EmailFromJSon{};
			e.Content = emailsForUser[i]["content"];
			e.Headers = make(map[string]string)
			for k,v := range emailsForUser[i] {
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
	var err = http.ListenAndServe(utilities.RECEIVE_BACKEND_LISTENS_FOR_MTA, mux);
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
			var users = ths.store.GetJSonBlob(map[string]string{"email": mailHeaders[model.EML_HDR_TO]});
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