package sendgridprovider
import (
"mail.bitlab.dk/model"
"mail.bitlab.dk/mtacontainer"
	"log"
)


type SendGridProvider struct {
	incoming chan model.Email;
	outgoing chan model.Email;
	events chan mtacontainer.Event;

}


func (ths *SendGridProvider) Stop() {

}

func (ths *SendGridProvider) GetOutgoing() chan model.Email{
	return ths.outgoing;
}

func (ths *SendGridProvider) GetIncoming() chan model.Email{
	return ths.incoming;
}

func (ths * SendGridProvider) GetEvent() chan mtacontainer.Event {
	return ths.events;
}

func (ths * SendGridProvider) GetName() string {
	return "Send Grid Email Service provider through API"
}



func New(log * log.Logger) mtacontainer.MTAProvider {
	var result = new(SendGridProvider);
	result.incoming = make(chan model.Email);
	result.outgoing = make(chan model.Email);
	result.events = make(chan mtacontainer.Event);
	log.Println(result.GetName()+ " MTA Going up")

	return result;
}