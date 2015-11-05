package mandrill
import (
"mail.bitlab.dk/mtacontainer"
"mail.bitlab.dk/model"
	"log"
)


type MandrillProvider struct {

	incoming chan model.Email;
	outgoing chan model.Email;
	events chan mtacontainer.Event;


}

func (ths *MandrillProvider) Stop() {

}

func (ths *MandrillProvider) GetOutgoing() chan model.Email{
	return ths.outgoing;
}

func (ths *MandrillProvider) GetIncoming() chan model.Email{
	return ths.incoming;
}

func (ths * MandrillProvider) GetEvent() chan mtacontainer.Event {
	return ths.events;
}

func (ths * MandrillProvider) GetName() string {
	return "Mandrill Email Service provider through API"
}

func New(log *log.Logger) mtacontainer.MTAProvider {
	var result = new(MandrillProvider);
	result.incoming = make(chan model.Email);
	result.outgoing = make(chan model.Email);
	result.events = make(chan mtacontainer.Event);
	log.Println(result.GetName()+ " MTA Going up")
	return result;
}


func (ths *MandrillProvider) sendingFunction() {
	for {


		select {
		case mail := <- ths.outgoing:
			println(mail);
		}

	}
}