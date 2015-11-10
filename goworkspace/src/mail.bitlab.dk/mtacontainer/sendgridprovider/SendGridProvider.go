//
// The SendGrid api used to implement this provider can be
// found here https://github.com/sendgrid/sendgrid-go.
//
//
// Author: Rasmus Winther Zakarias
//


package sendgridprovider
import (
	"mail.bitlab.dk/model"
	"mail.bitlab.dk/mtacontainer"
	"log"
	"mail.bitlab.dk/utilities/commandprotocol"
	"github.com/sendgrid/sendgrid-go"
	"errors"
	"strings"
	"mail.bitlab.dk/utilities"
	"mail.bitlab.dk/utilities/go"
)


const (
	SG_CNF_API_USER = "apiuser";
	SG_CNF_ENC_API_KEY = "encapikey";
	SG_CNF_API_KEY_LEN = "apikeylen";
	SG_CNF_PASSPHRASE = "passphrase";
)

type SendGridProvider struct {
	incoming        chan model.Email;
	outgoing        chan model.Email;
	events          chan mtacontainer.Event;
	cmd             chan commandprotocol.Command;
	sg              *sendgrid.SGClient;
	failureStrategy mtacontainer.FailureStrategy;
	log             *log.Logger;
}


func (ths *SendGridProvider) Stop() {
	ths.cmd <- commandprotocol.CMD_MTA_PROVIDER_SHUTDOWN;
}

func (ths *SendGridProvider) GetOutgoing() chan model.Email {
	return ths.outgoing;
}

func (ths *SendGridProvider) GetIncoming() chan model.Email {
	return ths.incoming;
}

func (ths * SendGridProvider) GetEvent() chan mtacontainer.Event {
	return ths.events;
}

func (ths * SendGridProvider) GetName() string {
	return "Send Grid Email Service "
}



func New(log *log.Logger, config map[string]string, fs mtacontainer.FailureStrategy) mtacontainer.MTAProvider {
	var result = new(SendGridProvider);
	result.incoming = make(chan model.Email);
	result.outgoing = make(chan model.Email);
	result.events = make(chan mtacontainer.Event);
	result.cmd = make(chan commandprotocol.Command);
	result.log = log;
	result.log.Println(result.GetName() + " MTA Going up")


	var apiKey=utilities.DecryptApiKey(config[SG_CNF_PASSPHRASE],config[SG_CNF_ENC_API_KEY],
		goh.StrToInt(config[SG_CNF_API_KEY_LEN]))
	result.sg = sendgrid.NewSendGridClientWithApiKey(apiKey);

	result.failureStrategy = fs;

	go result.sendingRoutine();
	return result;
}

// ---------------------------------------------------------
//
// SendGrid send routine
//
// ---------------------------------------------------------

// Go routine sending e-mails
func (ths *SendGridProvider) sendingRoutine() {

	for {
		select {
		case cmd := <-ths.cmd:
			if (cmd == commandprotocol.CMD_MTA_PROVIDER_SHUTDOWN) {
				ths.cmd <- commandprotocol.CMD_MTA_PROVIDER_NOTIFY_DOWN;
				return;
			}
		case email := <-ths.outgoing:
			ths.sgSend(email);
		}
	}

}

func (ths *SendGridProvider) sgSend(m model.Email) {


	message := sendgrid.NewMail()
	message.From = m.GetHeader(model.EML_HDR_FROM);
	message.To = m.GetHeaders()[model.EML_HDR_TO];
	message.Subject = m.GetHeader(model.EML_HDR_SUBJECT)
	for k, _ := range m.GetHeaders() {
		if strings.Compare(k,model.EML_HDR_FROM) != 0 &&
		   strings.Compare(k,model.EML_HDR_TO) != 0 &&
		   strings.Compare(k,model.EML_HDR_SUBJECT) != 0 {
			message.AddHeader(k, m.GetHeader(k));
		}
	}
	message.SetText(m.GetContent());

	err := ths.sg.Send(message)

	// report SendGrid Provider as down
	if err != nil {
		ths.log.Println(err.Error())
		if (ths.failureStrategy.Failure(mtacontainer.EK_CRITICAL) == false) {
			ths.events <- mtacontainer.NewEvent(mtacontainer.EK_RESUBMIT, errors.New("Sendgrid resubmits mail"),m);
		} else {
			ths.Stop(); // we are officially going down
			ths.events <- mtacontainer.NewEvent(mtacontainer.EK_FATAL, err);
			ths.log.Println("The SendGrid Provider is considered Down.");
			ths.events <- mtacontainer.NewEvent(mtacontainer.EK_RESUBMIT, errors.New("SendGrid is down for sending"), m);
			for e := range ths.outgoing {
				var ee model.Email = e;
				ths.events <- mtacontainer.NewEvent(mtacontainer.EK_RESUBMIT, errors.New("SendGrid is down for sending"), ee);
			}


		}
	} else {
		ths.failureStrategy.Success();
		ths.events <- mtacontainer.NewEvent(mtacontainer.EK_BEAT,
			errors.New("SendGrid send a message"));

	}
}