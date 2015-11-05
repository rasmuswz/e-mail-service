package smtpprovider
import (
"mail.bitlab.dk/mtacontainer"
"mail.bitlab.dk/model"
	"log"
	"mail.bitlab.dk/utilities/go"
)


type SMTPProvider struct {
	incoming chan model.Email;
	outgoing chan model.Email;
	events chan mtacontainer.Event;

	hostname string;
	port int;

}


func (ths *SMTPProvider) Stop() {

}

func (ths *SMTPProvider) GetOutgoing() chan model.Email{
	return ths.outgoing;
}

func (ths *SMTPProvider) GetIncoming() chan model.Email{
	return ths.incoming;
}

func (ths * SMTPProvider) GetEvent() chan mtacontainer.Event {
	return ths.events;
}

func (ths * SMTPProvider) GetName() string {
	return "SMTP ["+ths.hostname+":"+goh.IntToStr(ths.port)+"] Email Service provider through API"
}



func New(log * log.Logger, hostname string, port int) mtacontainer.MTAProvider {
	var result = new(SMTPProvider);
	result.incoming = make(chan model.Email);
	result.outgoing = make(chan model.Email);
	result.events = make(chan mtacontainer.Event);
	result.hostname = hostname;
	result.port = port;
	log.Println(result.GetName()+ " MTA Going up")
	return result;
}