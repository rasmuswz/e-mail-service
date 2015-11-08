//
// The LoopBackProvider reverses the the To and From fields of
// out going emails and routes them back through the {incoming} channel.
// E.g. sent e-mail are thrown back into the INBOX of the sender with
// TO and FROM reversed (if such an MBox where implemented).
//
// This is particularly useful when End-To-End testing.
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
	incoming        chan model.Email;
	outgoing        chan model.Email;
	events          chan mtacontainer.Event;
	command         chan commandprotocol.Command;
	failureStrategy mtacontainer.FailureStrategy;
}


func (ths *LoopBackProvider) Stop() {
	ths.command <- commandprotocol.CMD_MTA_PROVIDER_SHUTDOWN;
}

func (ths *LoopBackProvider) GetOutgoing() chan model.Email {
	return ths.outgoing;
}

func (ths *LoopBackProvider) GetIncoming() chan model.Email {
	return ths.incoming;
}

func (ths * LoopBackProvider) GetEvent() chan mtacontainer.Event {
	return ths.events;
}

func (ths * LoopBackProvider) GetName() string {
	return "Loop Back MTA Provider";
}



func New(log *log.Logger, fs mtacontainer.FailureStrategy) mtacontainer.MTAProvider {

	if fs == nil {
		return nil;
	}

	if log == nil {
		return nil;
	}

	var result = new(LoopBackProvider);
	result.incoming = make(chan model.Email);
	result.outgoing = make(chan model.Email);
	result.events = make(chan mtacontainer.Event);
	result.command = make(chan commandprotocol.Command);
	result.failureStrategy = fs;
	go result.handleOutgoingMessages();
	return result;
}



func (ths *LoopBackProvider) handleOutgoingMessages() {
	log.Println("Entering handleOutgoingMessages.");
	for {
		select {

		case m := <-ths.outgoing:
			log.Println("An outgoing event arrived.")
			if ths.failureStrategy.Failure(mtacontainer.EK_OK) == true {
				//Oh no we fail
				ths.events <- mtacontainer.NewEvent(mtacontainer.EK_RESUBMIT, errors.New("Loop Back MTA is failing and going down"), m);
				ths.events <- mtacontainer.NewEvent(mtacontainer.EK_FATAL, errors.New("Loop Back MTA is down"), ths);
				return; // <- This actually stops this provider :-)
			};

			ths.events <- mtacontainer.NewEvent(mtacontainer.EK_OK, errors.New("Loop Back MTA Got message"), ths)

			var headers = m.GetHeaders();
			if headers != nil {
				var temp = headers[model.EML_HDR_FROM];
				headers[model.EML_HDR_FROM] = headers[model.EML_HDR_TO];
				headers[model.EML_HDR_TO] = temp;
			}
			ths.incoming <- m;


		case c := <-ths.command:
			log.Println("Got a command");
			if c == commandprotocol.CMD_MTA_PROVIDER_SHUTDOWN {
				ths.events <- mtacontainer.NewEvent(mtacontainer.EK_FATAL, errors.New("Going down, SHUT DOWN Command"),
					ths);
				return;
			}
		}
	}
}