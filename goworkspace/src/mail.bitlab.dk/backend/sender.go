package backend
import (
	"mail.bitlab.dk/model"
	"net/http"
	"encoding/json"
	"log"
	"mail.bitlab.dk/utilities"
	"bytes"
)

type SendBackEnd struct {
	store    JSonStore;
	outgoing chan model.Email;
	cmd chan int;
}

func (ths * SendBackEnd) Stop() {
	ths.cmd <- CMD_SHUTDOWN;
}

// ------------------------------------------------------
//
// The ClientAPI connects to send an e-mail, we reads the
// mail and forwards it to the MTAContainer.
//
// ------------------------------------------------------
func (ths *SendBackEnd) sendmail(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close();
	var _, ok = CheckAuthorizedUser(ths.store, r);
	if ok == false {
		http.Error(w,"Access denied",http.StatusForbidden);
		return;
	}

	var email = model.EmailImpl{};

	err := utilities.ReadJSonBody(r,&email);
	if err != nil {
		http.Error(w,err.Error(),http.StatusInternalServerError);
		return;
	}

	ths.outgoing <- &email;
}

func (ths *SendBackEnd) ListenForClientApiSendingMails() {
	var mux = http.NewServeMux();
	mux.HandleFunc("/sendmail", ths.sendmail);
	http.ListenAndServe(utilities.SEND_BACKEND_LISTEN_FOR_CLIENT_API, mux);
}

func (ths *SendBackEnd) ForwardToMtaContainer() {
	for {
		select {

		case email := <-ths.outgoing:
			var client = new(http.Client);

			var mailData, mailDataErr = json.Marshal(&email);

			if mailDataErr != nil {
				log.Println("[SendBackEnd] Failed to marshall email:" + mailDataErr.Error());
				return;
			}

			var resp, respErr = client.Post(utilities.MTA_LISTENS_FOR_SEND_BACKEND, "text/json", bytes.NewReader(mailData));
			if respErr != nil {
				log.Println("[SendBackEnd] Failed to communicate with MTA Container: " + respErr.Error());
			} else {
				log.Println("[SendBackEnd] Mail Transferred to Mta Container:");
			}
			resp.Body.Close();

		case cmd := <-ths.cmd:
			if (cmd == CMD_SHUTDOWN) {
				log.Println("[Sender BackEnd] Good bye");
				return;
			}
		}
	}
}


func NewSendBackend(store JSonStore) *SendBackEnd{
	var result *SendBackEnd = new (SendBackEnd);
	result.store = store;
	result.outgoing = make(chan model.Email);
	result.cmd = make(chan int);
	go result.ListenForClientApiSendingMails();
	go result.ForwardToMtaContainer();
	return result;
}

//
// Takes the calling thread for service.
//
func RunSendBackEnd(store JSonStore) {
	var result *SendBackEnd = new (SendBackEnd);
	result.store = store;
	result.outgoing = make(chan model.Email);
	go result.ListenForClientApiSendingMails();
	result.ForwardToMtaContainer();
}