//
// Welcome to the MailGun MTA Provider documentation
// for the thirdparty API in mailgun-go can be found
// here https://www.github.com/mailgun/mailgun-go.
//
// Author: Rasmus Winther Zakarias
//
//
package mailgunprovider
import (
	"mail.bitlab.dk/model"
	"log"
	"github.com/mailgun/mailgun-go"
	"errors"
	"mail.bitlab.dk/mtacontainer"
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
// We implement the MailGun provider as a Go-routine
// - one for sending (out)
//
// Messages to be sent are posted on the {out} channel.
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
	events chan mtacontainer.Event;
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
	return m.events;
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

	//m.sendMaintanenceMessage("Admin,\nPlease find the MailGun MTA provider is going **down**.");

	// wait for both receiving routine and sending routine to Stop.
	var c = <-m.cmd;
	if (c != CMD_RECV_HAS_TERMINATED && c != CMD_SEND_HAS_TERMINATED) {
		m.log.Println("Unknown message on Cmd channel during shutdown? (Review code)");
	}

//	c = <-m.cmd;
//	if (c != CMD_RECV_HAS_TERMINATED && c != CMD_SEND_HAS_TERMINATED) {
//		m.log.Println("Unknown message on Cmd channel during shutdown? (Review code)");
//	}

	// make clients know this instance is gone.
	close(m.events);
	close(m.cmd);
	close(m.inc);
	close(m.out);
}


// ---------------------------------------------------------------
//
// Implementation
//
// ---------------------------------------------------------------
func New(log *log.Logger, config map[string]string, fs mtacontainer.FailureStrategy) mtacontainer.MTAProvider {
	var result *MailGunProvider = new(MailGunProvider);

	result.failureStrategy = fs;

	// -- initialize the result --
	result.config = config;
	var apiKey = utilities.DecryptApiKey(config[MG_CNF_PASSPHRASE], config[MG_CNF_ENCRYPTED_APIKEY],
		goh.StrToInt(config[MG_CNF_ENCRYPTED_APIKEY_LEN]));

	result.mg = mailgun.NewMailgun(config[MG_CNF_DOMAIN_TO_SERVE], apiKey, "");
	result.log = log;

	// setup channels
	result.out = make(chan model.Email);
	result.inc = make(chan model.Email);
	result.cmd = make(chan Cmd);
	result.events = make(chan mtacontainer.Event);

	// for routines to serve
	go result.sendingRoutine();
	//go result.receivingRoutine();

	log.Println(result.GetName() + " MTA Going up")

	// send initial "MG is in the Air"-message to Admin
	//result.sendMaintanenceMessage("Admin,\nPlease find the MailGun MTA provider is up.");

	return result;
}


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

	from := m.GetHeader(model.EML_HDR_FROM);
	to := m.GetHeader(model.EML_HDR_TO);
	subject := m.GetHeader(model.EML_HDR_SUBJECT);
	content := m.GetContent();

	message := mgp.mg.NewMessage(from,subject,content,to);

	mgp.log.Println("Invoking MailGun API to send e-mail");
	var mm, mailId, err = mgp.mg.Send(message);

	// report MailGunProvider as down
	if err != nil {
		mgp.log.Println("MG Failed to send e-mail");
		mgp.log.Println(err.Error())
		if (mgp.failureStrategy.Failure(mtacontainer.EK_CRITICAL) == false) {
			mgp.events <- mtacontainer.NewEvent(mtacontainer.EK_WARNING,err,mgp);
			mgp.events <- mtacontainer.NewEvent(mtacontainer.EK_RESUBMIT, err,m);
		} else {
			mgp.Stop(); // we are officially going down
			mgp.events <- mtacontainer.NewEvent(mtacontainer.EK_FATAL, err);
			mgp.log.Println("The MailGun Provider is considered Down.");
			mgp.events <- mtacontainer.NewEvent(mtacontainer.EK_RESUBMIT,errors.New("MailGun is down for sending"),m);
			for e := range mgp.out {
				var ee model.Email = e;
				mgp.events <- mtacontainer.NewEvent(mtacontainer.EK_RESUBMIT,errors.New("MailGun is down for sending"),ee);
			}


		}
	} else {
		mgp.log.Println("MG Has sent email with success.");
		mgp.failureStrategy.Success();
		mgp.events <- mtacontainer.NewEvent(mtacontainer.EK_BEAT,
			errors.New("MailGun says: " + mm + " for sending message giving it id " + mailId));
		mgp.log.Println("After sending event on event channel");

	}
}
