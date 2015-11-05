//
// The LoopBackProvider reverses the the To and From fields of
// out going emails and delivers them to the Sender, that is Here.
//
//
// Author: Rasmus Winther Zakarias
//
package loopbackprovider
import (
	"mail.bitlab.dk/model"
	"mail.bitlab.dk/mtacontainer"
	"log"
	"mail.bitlab.dk/utilities/commandprotocol"
	"errors"
)





type LoopBackProvider struct {
	incoming chan model.Email;
	outgoing chan model.Email;
	events chan mtacontainer.Event;
	command chan commandprotocol.Command;
}


func (ths *LoopBackProvider) Stop() {

}

func (ths *LoopBackProvider) GetOutgoing() chan model.Email{
	return ths.outgoing;
}

func (ths *LoopBackProvider) GetIncoming() chan model.Email{
	return ths.incoming;
}

func (ths * LoopBackProvider) GetEvent() chan mtacontainer.Event {
	return ths.events;
}

func (ths * LoopBackProvider) GetName() string {
	return "Loop Back MTA Provider";
}



func New(log * log.Logger) mtacontainer.MTAProvider {
	var result = new(LoopBackProvider);
	result.incoming = make(chan model.Email);
	result.outgoing = make(chan model.Email);
	result.events = make(chan mtacontainer.Event);
	log.Println(result.GetName()+ " MTA Going up")
	go result.handleOutgoingMessages();
	return result;
}



func (ths *LoopBackProvider) handleOutgoingMessages() {

	for{
		select {

		case m := <-ths.incoming:

		var headers = m.GetHeaders();
		var temp = headers[model.EML_HDR_FROM];
		headers[model.EML_HDR_FROM] = headers[model.EML_HDR_TO];
		headers[model.EML_HDR_TO] = temp;
			ths.outgoing <- m;


		case c := <-ths.command:
			if c == commandprotocol.CMD_MTA_PROVIDER_SHUTDOWN {
				ths.events <- mtacontainer.NewEvent(mtacontainer.EK_FATAL,errors.New("Going down, SHUT DOWN Command"),
					ths);
				return;
			}
		}
	}
}