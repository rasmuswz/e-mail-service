//
// Welcome to the MailGun MTA Provider.
//
//
//
//
//
//
package mailgunprovider
import (
	"mail.bitlab.dk/model"
	"log"
	"github.com/mailgun-go"
	"net/http"
	"strconv"
	"errors"
	"io/ioutil"
	"time"
	"mail.bitlab.dk/mtacontainer"
	"os"
	"strings"
)

const (
// The domain what we serve mails for
	MG_SRV_DMN = "mail.bitlab.dk";
// REST service access point
	MG_API_URL = "https://api.mailgun.net/v3/" + MG_SRV_DMN + "/";
// Key gotten from MG-account using log-in at
// https://mailgun.com/sessions/new
	MG_API_KEY = "key-1693daa393947a8e56e33133b82ca1cd";
// E-mail to notify when service state changes e.g. when it goes up and down
	MG_RPT_EML = "r@wz.gl"; // Send health information about this provider to this address
// Route Action that account MailGun shall have
	MG_RUT_ACT="forward(\"https://mail.bitlab.dk:31415/msg\")";
)



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
	var headers = m.GetHeaders();

	var message = mgp.mg.NewMessage(headers[model.EML_HDR_FROM],
		headers[model.EML_HDR_SUBJECT],
		m.GetContent(),
		model.EML_HDR_TO);
	var mm, mailId, err = mgp.mg.Send(message);


	// report MailGunProvider as down
	if err != nil {
		mgp.log.Println(err.Error())
		mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_DOWN, err);
	} else {
		mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_BEAT,
			errors.New("MailGun says: "+mm+" for sending message giving it id "+mailId));

	}
}

// ---------------------------------------------------------------
// Receiving E-mails and Health information (WebHooks)
// ---------------------------------------------------------------

const MG_SERVE_AT_MSG = "MailGun server at " + MG_API_URL;

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
	err := http.ListenAndServeTLS(":31415", "cert.pem", "key.pem", mux);

	if (err != nil) {
		d, e := os.Getwd();
		if (e == nil) {
			mgp.log.Println("We are looking for: " + d + "/cert.pem "); }
		mgp.log.Println("Error: " + err.Error());
		mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_DOWN, err);
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
		mgp.log.Println("Error: " + MG_SERVE_AT_MSG + " failed reporting routes:\n" + err.Error() + "\n" +
		"Terminating receiving routine.");
		mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_DOWN, err);
		return false;
	}

	//
	// Also the service is down if the route is not found
	//
	mgp.log.Println(MG_SERVE_AT_MSG + " reports " + strconv.Itoa(routesOnServer) + " available");
	routeFound := false;
	for r := range routes {
		route := routes[r];
		for a := range route.Actions {
			action := route.Actions[a];
			if (strings.Compare(action,MG_RUT_ACT) == 0) {
				routeFound = true;
				mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_OK,
					errors.New("Everything is fine, we found the route."));
			}
		}
	}
	if (routeFound == false) {
		mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_DOWN,
			errors.New("forward(\"https://mail.bitlab.dk:31415/msg\") route not found"));
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
		mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_DOWN, err);
		return nil;
	}

	if (hooks[hook] == "") {
		err = mgp.mg.CreateWebhook(hook, "https://mail.bitlab.dk:31415/" + hook);
		if (err != nil) {
			mgp.health <- mtacontainer.NewEvent(mtacontainer.EK_DOWN, err);
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

	var incoming = model.NewEmail(bodyString, model.EML_HDR_TO, r.Header[model.EML_HDR_TO][0], model.EML_HDR_FROM, r.Header[model.EML_HDR_FROM][0]);

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
	var goingUpMessage = model.NewEmail(msg,
		model.EML_HDR_FROM, "mailgun@mail.bitlab.dk",
		model.EML_HDR_TO, "r@wz.gl",
		model.EML_HDR_SUBJECT, "[MailGun] Provider starting up");
	mgp.out <- goingUpMessage;
}

// Construct a Mail Gun Provider
func NewMailGun(log *log.Logger) mtacontainer.MTAProvider {
	var result *MailGunProvider = new(MailGunProvider);
	result.mg = mailgun.NewMailgun("mail.bitlab.dk", MG_API_KEY, "");
	result.log = log;

	// setup channels
	result.out   = make(chan model.Email);
	result.inc   = make(chan model.Email);
	result.cmd   = make(chan Cmd);
	result.health= make(chan mtacontainer.Event);

	// for routines to serve
	go result.sendingRoutine();
	go result.receivingRoutine();

	// send initial "MG is in the Air"-message to Admin
	//result.sendMaintanenceMessage("Admin,\nPlease find the MailGun MTA provider up.");

	return result;
}