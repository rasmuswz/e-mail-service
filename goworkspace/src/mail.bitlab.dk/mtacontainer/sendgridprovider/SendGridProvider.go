package sendgridprovider
import (
	"mail.bitlab.dk/model"
	"mail.bitlab.dk/mtacontainer"
	"log"
	"mail.bitlab.dk/utilities/commandprotocol"
	"github.com/sendgrid/sendgrid-go"
	"errors"
)


const (
	SG_CNF_API_USER = "apiuser";
	SG_CNF_API_KEY = "apikey";

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
	return "Send Grid Email Service provider through API"
}



func New(log *log.Logger, config map[string]string) mtacontainer.MTAProvider {
	var result = new(SendGridProvider);
	result.incoming = make(chan model.Email);
	result.outgoing = make(chan model.Email);
	result.events = make(chan mtacontainer.Event);
	result.log = log;
	result.log.Println(result.GetName() + " MTA Going up")
	result.sg = sendgrid.NewSendGridClient(config[SG_CNF_API_USER], config[SG_CNF_API_KEY]);
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
	for k, _ := range m.GetHeaders() {
		message.AddHeader(k, m.GetHeader(k));
	}
	message.SetText(m.GetContent());

	err := ths.sg.Send(message)

	// report MailGunProvider as down
	if err != nil {
		ths.log.Println(err.Error())
		if (ths.failureStrategy.Failure(mtacontainer.EK_CRITICAL) == false) {
			ths.events <- mtacontainer.NewEvent(mtacontainer.EK_DOWN_TEMPORARILY, err);
		} else {
			ths.Stop(); // we are officially going down
			ths.events <- mtacontainer.NewEvent(mtacontainer.EK_FATAL, err);
			ths.log.Println("The MailGun Provider is considered Down.");
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