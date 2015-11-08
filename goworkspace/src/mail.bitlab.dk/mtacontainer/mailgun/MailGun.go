//
// Welcome to the MailGun MTA Provider.
//
//
//
package mailgunprovider
import (
	"mail.bitlab.dk/model"
	"log"
	"github.com/mailgun/mailgun-go"
	"net/http"
	"strconv"
	"errors"
	"io/ioutil"
	"time"
	"mail.bitlab.dk/mtacontainer"
	"os"
	"strings"

	"mail.bitlab.dk/utilities"
	"mail.bitlab.dk/utilities/go"
)

const (
	// The map[string]string MG configuration has the following defined keys
	MG_CNF_PASSPHRASE = "apikeykey";
	MG_CNF_ENCRYPTED_APIKEY = "encapikey";
	// stiputales the length of the plaintext key.
	MG_CNF_ENCRYPTED_APIKEY_LEN = "apikeylen";
	MG_CNF_ROUTE_ACTION_ON_INCOMING_MAIL = "routeaction";
	MG_CONF_HEALTH_NOTIFY_EMAIL = "notifymail";
	MG_CNF_DOMAIN_TO_SERVE = "domain";
)

func (ths *MailGunProvider) createMGApiUrl(domain_to_serve string ) string{
	return "https://api.mailgun.net/v3/" + domain_to_serve + "/";
}

// ---------------------------------------------------------------
// MailGunProvider implementation
//
// We implement the MailGun provider as two Go-routines
// - one for sending (out)
// - one for receiving (inc)
//
// A new MailGun provider spawns two go routines one for
// receiving mails posting on channel {inc} and one for
// sending mails consuming from the {out} channel.
//
// To control the MailGun provider we provide a {cmd}
// channel. Errors and HeartBeats are reported as events
// on the {health} channel.
//
// ---------------------------------------------------------------
type MailGunProvider struct {
	log    *log.Logger;
	mg     mailgun.Mailgun;
	cmd    chan Cmd;
	out    chan model.Email;
	inc    chan model.Email;
	health chan mtacontainer.Event;
	config map[string]string;
	failureStrategy mtacontainer.FailureStrategy;
}



// Commands to control this service
type Cmd uint32;

// The Cmd channel has these three messages
const (
// Tell Send and Receive routes to die
	CMD_TERMINATE = 0x01;
// Report that Send has terminated
	CMD_SEND_HAS_TERMINATED = 0x02;
// Report that Receive has terminated
	CMD_RECV_HAS_TERMINATED = 0x03;
)


// ---------------------------------------------------------------
//
// public API
//
// ---------------------------------------------------------------
func (m *MailGunProvider) GetEvent() chan mtacontainer.Event {
	return m.health;
}

func (m *MailGunProvider) GetIncoming() chan model.Email {
	return m.inc;
}

func (m *MailGunProvider) GetOutgoing() chan model.Email {
	return m.out;
}

func (m *MailGunProvider) GetName() string {
	return "Mail Gun Provider";
}

func (m *MailGunProvider) Stop() {
	m.cmd <- CMD_TERMINATE;

	m.sendMaintanenceMessage("Admin,\nPlease find the MailGun MTA provider is going **down**.");

	// wait for both receiving routine and sending routine to Stop.
	var c = <-m.cmd;
	if (c != CMD_RECV_HAS_TERMINATED && c != CMD_SEND_HAS_TERMINATED) {
		m.log.Println("Unknown message on Cmd channel during shutdown? (Review code)");
	}
	c = <-m.cmd;
	if (c != CMD_RECV_HAS_TERMINATED && c != CMD_SEND_HAS_TERMINATED) {
		m.log.Println("Unknown message on Cmd channel during shutdown? (Review code)");
	}

	// make clients know this instance is gone.
	close(m.health);
	close(m.cmd);
	close(m.inc);
	close(m.out);
}


// ---------------------------------------------------------------
//
// Implementation
//
// ---------------------------------------------------------------


// ---------------------------------------------------------------
// Sending e-mails
// ---------------------------------------------------------------

// Go routine sending e-mails
func (mgp *MailGunProvider) sendingRoutine() {

	for {
		select {
		case cmd := <-mgp.cmd:
			if (cmd == CMD_TERMINATE) {
				mgp.cmd <- CMD_SEND_HAS_TERMINATED;
				return;
			}
		case email := <-mgp.out:
			mgp.mgSend(email);
		}
	}

}

func (mgp *MailGunProvider) mgSend(m model.Email) {




	var message = mgp.mg.NewMessage(m.GetHeader(model.EML_HDR_FROM),
		m.GetHeader(model.EML_HDR_SUBJECT),
		m.GetContent(),
		m.GetHeader(model.EML_HDR_TO));
	for k,_ := range m.GetHeaders() {
		if k != model.EML_HDR_SUBJECT && k != model.EML_HDR_FROM {
			message.AddHeader(k, m.GetHeader(k));
		}
	}

	var mm, mailId, err = mgp.mg.Send(message);


	// report MailGunProvider as down
	if err != nil {
		mgp.log.Println(err.Error())
		if (mgp.failureStrategy.Failure(mtacontainer.EK_CRITICAL) == false) {
			mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_DOWN_TEMPORARILY, err);
		} else {
			mgp.Stop(); // we are officially going down
			mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_FATAL, err);
			mgp.log.Println("The MailGun Provider is considered Down.");
			mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_RESUBMIT,errors.New("MailGun is down for sending"),m);
			for e := range mgp.out {
				var ee model.Email = e;
				mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_RESUBMIT,errors.New("MailGun is down for sending"),ee);
			}


		}
	} else {
		mgp.failureStrategy.Success();
		mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_BEAT,
			errors.New("MailGun says: " + mm + " for sending message giving it id " + mailId));

	}
}

// ---------------------------------------------------------------
// Receiving E-mails and Health information (WebHooks)
// ---------------------------------------------------------------



//
// MailGun provides Routes to notify external service of incoming mails.
// See https://documentation.mailgun.com/api-routes.html.
//
// We check at at least one route is set up in the MailGun account
// and listens for at http://mail.bitlab.dk:31514/msg.
//
// We try to check whether a forward route for is present
// http://mail.bitlab.dk:31514/msg and report a health problem
// is this is not the case. E.g. warning about not receiving mails
// delivered at MailGun. For our Free account option this is severe
// as unhandled mails are erased af two days.
//
func (mgp *MailGunProvider) receivingRoutine() {

	// Check that a route is set for us
	var routeChecksOut = mgp.checkRoute();
	if routeChecksOut == false {
		return;
	}

	// TODO(rwz): ReFactor, separate receiving e-mails from handling WebHooks

	// open service at http://mail.bitlab.dk:31514/ with hooks and for incoming mail
	var mux = http.NewServeMux();
	mux.Handle("/msg", newMgIncomingHandler(mgp));
	mux.Handle("/deliver", mgp.checkAndAddHook("deliver", mgp.logHandler));
	mux.Handle("/drop", mgp.checkAndAddHook("drop", mgp.logHandler));
	mux.Handle("/spam", mgp.checkAndAddHook("spam", mgp.logHandler));
	mux.Handle("/unsubscribe", mgp.checkAndAddHook("unsubscribe", mgp.logHandler));
	mux.Handle("/click", mgp.checkAndAddHook("click", mgp.logHandler));
	mux.Handle("/open", mgp.checkAndAddHook("open", mgp.logHandler));
	err := http.ListenAndServeTLS(utilities.MTA_MAILGUN_SERVICE_PORT, "cert.pem", "key.pem", mux);

	if (err != nil) {
		d, e := os.Getwd();
		if (e == nil) {
			mgp.log.Println("We are looking for: " + d + "/cert.pem "); }
		mgp.log.Println("Error: " + err.Error());
		mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_DOWN_TEMPORARILY, err);
	}

	mgp.log.Println("MailGun provider online.");

	for {
		select {
		case cmd := <-mgp.cmd:
			if cmd == CMD_TERMINATE {
				mgp.cmd <- CMD_RECV_HAS_TERMINATED;
				return;
			}
		}
	}
}


//
// Check Route
//
// Download all Routes for the mailgun-account and find one that
// forwards all mails for mail.bitlab.dk to
// https://mail.bitlab.dk:31415/msg
//
//
func (mgp *MailGunProvider) checkRoute() bool {

	//
	// Lacking documentation in mailgun.go added
	// from https://documentation.mailgun.com/api-routes.html#actions
	//
	// Fetches the list of routes. Note that routes are defined globally,
	// per account, not per domain as most of other API calls.
	//
	// param limit - Maximum number of records to return. (100 by default)
	// param skip  - Number of records to skip. (0 by default)
	//
	// Note(rwz): We assume here that the Route we search for is
	// within the first 100 routes.
	//
	var routesOnServer, routes, err = mgp.mg.GetRoutes(mailgun.DefaultLimit, mailgun.DefaultSkip);

	//
	// Report problems invoking MailGun as service is DOWN, EK_DOWN.
	//
	if err != nil {
		mgp.log.Println("Error: service failed in reporting routes:\n" + err.Error() + "\n" +
		"Terminating receiving routine.");
		mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_DOWN_TEMPORARILY, err);
		return false;
	}

	//
	// Also the service is down if the route is not found
	//
	mgp.log.Println(mgp.createMGApiUrl(mgp.config[MG_CNF_DOMAIN_TO_SERVE]) +
	  " reports " + strconv.Itoa(routesOnServer) + " available");
	routeFound := false;
	for r := range routes {
		route := routes[r];
		for a := range route.Actions {
			action := route.Actions[a];
			if (strings.Compare(action, mgp.config[MG_CNF_ROUTE_ACTION_ON_INCOMING_MAIL]) == 0) {
				routeFound = true;
				mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_OK,
					errors.New("Everything is fine, we found the route."));
			}
		}
	}
	if (routeFound == false) {
		mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_DOWN_TEMPORARILY,
			errors.New("forward(\"https://mail.bitlab.dk:"+utilities.MTA_MAILGUN_SERVICE_PORT+"msg\") route not found"));
		return false;
	}

	return true;

}

func (m * MailGunProvider) logHandler(w http.ResponseWriter, r *http.Request) {
	m.log.Println(r.Host + " performing " + r.Method + " on path " + r.RequestURI);
}

func (mgp *MailGunProvider) checkAndAddHook(hook string, fn serveFn) http.Handler {

	hooks, err := mgp.mg.GetWebhooks();

	if (err != nil) {
		mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_DOWN_TEMPORARILY, err);
		return nil;
	}

	if (hooks[hook] == "") {
		err = mgp.mg.CreateWebhook(hook, "https://mail.bitlab.dk"+utilities.MTA_MAILGUN_SERVICE_PORT+"/" + hook);
		if (err != nil) {
			mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_DOWN_TEMPORARILY, err);
			// TODO(rwz): Consider whether this is the proper action.
		}
	}

	return NewFnHandler(fn);
}

type serveFn func(w http.ResponseWriter, r *http.Request);

type FnHandler struct {
	srv serveFn;
}

func (fnH *FnHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fnH.srv(w, r);
}

func NewFnHandler(f serveFn) http.Handler {
	var res = new(FnHandler);
	res.srv = f;
	return res;
}

type MGIncomingMessageHandler struct {
	mgp *MailGunProvider;
}

func (mgh *MGIncomingMessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var start = time.Now();
	mgh.mgp.log.Println("Incoming e-mail processing it, stand by ...");



	var rawBody, err = ioutil.ReadAll(r.Body);
	if (err != nil) {
		mgh.mgp.log.Println("Error: while retrieving mail from request: " + err.Error());
		return;
	}
	var bodyString = string(rawBody);

	var incoming = model.NewEmailFlattenHeaders(bodyString, model.EML_HDR_TO, r.Header[model.EML_HDR_TO][0], model.EML_HDR_FROM, r.Header[model.EML_HDR_FROM][0]);

	mgh.mgp.inc <- incoming;
	mgh.mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_BEAT, errors.New("Mail service is alive and delivered a message."));

	mgh.mgp.log.Println("Handling e-mail took: " + start.Sub(time.Now()).String());
}

func newMgIncomingHandler(mgp *MailGunProvider) http.Handler {
	var r = new(MGIncomingMessageHandler);
	r.mgp = mgp;
	return r;
}

func (mgp *MailGunProvider) sendMaintanenceMessage(msg string) {
	var goingUpMessage = model.NewEmailFlattenHeaders(msg,
		model.EML_HDR_FROM, "mailgun@mail.bitlab.dk",
		model.EML_HDR_TO, mgp.config[MG_CONF_HEALTH_NOTIFY_EMAIL],
		model.EML_HDR_SUBJECT, "[MailGun] Provider starting up");
	mgp.out <- goingUpMessage;
}



// Construct a Mail Gun Provider


func n(log *log.Logger, config map[string]string, fs mtacontainer.FailureStrategy) mtacontainer.MTAProvider {
	var result *MailGunProvider = new(MailGunProvider);

	result.failureStrategy = fs;

	// -- initialize the result --
	result.config = config;
	var apiKey = utilities.DecryptApiKey(config[MG_CNF_PASSPHRASE], config[MG_CNF_ENCRYPTED_APIKEY],
		goh.StrToInt(config[MG_CNF_ENCRYPTED_APIKEY_LEN]));

		print("MG: "+apiKey);

	result.mg = mailgun.NewMailgun(config[MG_CNF_DOMAIN_TO_SERVE], apiKey, "");
	result.log = log;

	// setup channels
	result.out = make(chan model.Email);
	result.inc = make(chan model.Email);
	result.cmd = make(chan Cmd);
	result.health = make(chan mtacontainer.Event);

	// for routines to serve
	go result.sendingRoutine();
	go result.receivingRoutine();

	// send initial "MG is in the Air"-message to Admin
	result.sendMaintanenceMessage("Admin,\nPlease find the MailGun MTA provider is up.");

	return result;
}

func New(log *log.Logger, config map[string]string, failureStrategy mtacontainer.FailureStrategy) mtacontainer.MTAProvider {
	return n(log,config,failureStrategy);
}