package mtacontainer
/**
 *
 * An MTA Container holds an array of MTAProviders. Each MTA Provider represents
 * the services offered by a given provider. E.g. listen of incoming mail and
 * and sending e-mails out.
 *
 * Author: Rasmus Winther Zakarias
 */

import (
	"mail.bitlab.dk/model"
	"time"
	"log"
	"mail.bitlab.dk/utilities"
	"errors"
	"encoding/json"
	"net/http"
	"mail.bitlab.dk/utilities/commandprotocol"
	"bytes"
"strings"
)

type EventKind uint32;
const (
// no problem
	EK_OK = 0x00;
// reporting entity is out of service but we will try again
// and hopefully recover.
	EK_DOWN_TEMPORARILY = 0x01;
// reporting entity suffered a time out
// but we haven't given up yet will try again
	EK_TIMEOUT = 0x02;
// reporting entity beats (it is up and well)
	EK_BEAT = 0x03;
// When the Error message holds a warning
// maintainers should know about
	EK_WARNING = 0x04;
// When the Error message holds a serious warning
// that may be cause interrupt service.
	EK_CRITICAL = 0x05;
// When an Error occurs severe enough that the reporting entity
// is considered unavailable and we have no means for recovering
// from this.
	EK_FATAL = 0x06;
// We failed, and Payload needs to be submitted else where.
	EK_RESUBMIT = 0x07;
	// We must inform the user
	EK_INFORM_USER = 0x08;
	// ShutDown container, this event releases the main go routine and tearsdown the application.
	EK_GOOD_BYE = 0x09;
)

// ------------------------------------------------------------------
//
// Failure Strategy
//
// Intended use is that a service experiencing a (non fatal) failure
// will invoke its FailureStrategy telling it whether it should wait
// and try later or die-hard.
// ------------------------------------------------------------------
type FailureStrategy interface {
	// returns true if patience has run out
	// false if no action should be taken.
	Failure(f EventKind) bool
	Success();
}

// ------------------------------------------------------------
//
// A Threshold Failure Strategy
//
// More than {threshold} consecutive failures and we die-hard.
//
// ------------------------------------------------------------
type ThresholdFailureStrategy struct {
	noSevereConsecutiveFailures int;
	threshold                   int;
}

func (ths * ThresholdFailureStrategy) Failure(f EventKind) bool {
	if (f > EK_WARNING) {
		ths.noSevereConsecutiveFailures += 1;
	}

	return ths.noSevereConsecutiveFailures > ths.threshold;
}


func (ths * ThresholdFailureStrategy) Success() {
	ths.noSevereConsecutiveFailures = 0;
}

func NewThressHoldFailureStrategy(thresshold int) FailureStrategy {
	result := new(ThresholdFailureStrategy);
	result.noSevereConsecutiveFailures = 0;
	result.threshold = thresshold;
	return result;
}

/**
 * Events from {HealthService} tells some kind of event happened with the
 * **STATE** of the system. It can merely be that a heart-beat was received
 * from a provider stipulating it is up, in which case the EK_OK event is
 * appropriate and {getError} will report nil.
 */
type Event interface {
	GetKind() EventKind;
	GetError() error;
	GetTime() time.Time;
	GetPayload() interface{};
}

/**
 *
 * Any component implementing this interface supplies health information
 * through its {Event} channel. The channel is intended to be on-way e.g.
 * only a service post Events on its Event channel, client code merely
 * observes these events.
 *
 */
type HealthService interface {
	GetEvent() chan Event;
}

type defaultEvent struct {
	time    time.Time;
	err     error;
	kind    EventKind;
	payload interface{};
}

func (de *defaultEvent) GetKind() EventKind {
	return de.kind;
}

func (de * defaultEvent) GetError() error {
	return de.err;
}

func (de * defaultEvent) GetTime() time.Time {
	return de.time;
}

func (de * defaultEvent) GetPayload() interface{} {
	return de.payload;
}

func NewEvent(kind EventKind, e error, sender ...interface{}) Event {
	var r *defaultEvent = new(defaultEvent);
	r.kind = kind;
	r.err = e;
	r.time = time.Now();
	if (len(sender) > 0) {
		r.payload = sender[0];
	}
	return r;
}

/*
 * Mail Transport Agent Services offers a channel for receiving
 * emails and a chanel for sending them. This project has focus on
 * sending e-mails but maybe one day we would like to receive them too.
 *
 * These is a {Stop} method to shutdown gracefully.
 *
 */
type MTAService interface {
	GetIncoming() chan model.Email;
	GetOutgoing() chan model.Email;
	Stop();
}


/*
 * When created an MTA Provider (depending on implementation) will
 * listen for the actual provider to supply emails. Whenever an e-mail
 * arrives the {mail}-channel will have an e-mail struct ready.
 */
type MTAProvider interface {
	MTAService;
	HealthService;
	GetName() string;
}



/*
 * The {MTAContainer} keeps track of a set of MTAProviders and provides an
 * unified interface for Email-service using a schedule for balancing the load
 * between the providers.
 *
 * Furthermore the MTAContainer provides a health service.
 */
type MTAContainer interface {
	GetProviders() []MTAProvider;
	SetScheduler(scheduler Scheduler);
	MTAService;
	HealthService;
}


// ------------------------------------------------------
//
// Default Mail Transport Agent Container implementation
//
// We maintain a list of MTAProviders and aggregate their
// incoming and health channels into those of the container.
//
// For outgoing mail we take from the contains outgoing channel
// run the scheduler selecting a provider and uses the selected
// provider for sending the message.
//
//
// ------------------------------------------------------
type DefaultMTAContainer struct {
	providers []MTAProvider;
	scheduler Scheduler;
	incoming  chan model.Email;
	outgoing  chan model.Email;
	events    chan Event;
	cmd       chan commandprotocol.Command;
	log       *log.Logger;
}


func (d *DefaultMTAContainer) GetProviders() []MTAProvider {
	return d.providers;
}

func (d *DefaultMTAContainer) SetScheduler(scheduler Scheduler) {
	d.scheduler = scheduler;
}

func (d *DefaultMTAContainer) GetIncoming() chan model.Email {
	return d.incoming;
}

func (d *DefaultMTAContainer) GetOutgoing() chan model.Email {
	return d.outgoing;
}

func (d *DefaultMTAContainer) GetEvent() chan Event {
	return d.events;
}

func (d *DefaultMTAContainer) Stop() {
	for i := range d.providers {
		var provider = d.providers[i];
		provider.Stop();
	}
	d.cmd <- commandprotocol.CMD_MTA_PROVIDER_SHUTDOWN;
}

//
// Public constructor for creating a DefaultMTAContainer.
//
func New(scheduler Scheduler) MTAContainer {
	if (scheduler == nil) {
		return nil;
	}

	var result = new(DefaultMTAContainer);
	result.providers = scheduler.GetProviders();
	result.incoming = make(chan model.Email);
	result.outgoing = make(chan model.Email);
	result.events = make(chan Event);
	result.log = utilities.GetLogger("MTAContainer");

	result.scheduler = scheduler;

	// Aggregate all provider incoming and event channels to those of the MTAContainer
	for p := range result.providers {
		var pp MTAProvider = result.providers[p];

		// incoming
		go func() {
			for {
				var in = pp.GetIncoming();
				var email, ok = <-in;  // suppose an MTA some day delivers an email it happens here
				if ok == false {
					result.log.Println("Incoming was closed.")
				}
				result.incoming <- email;
			}
		}();

		//
		// Manage MTAProviders
		//
		go func() {
			for {
				var evt = pp.GetEvent();
				var event, ok = <-evt;
				if (ok == false) {
					result.log.Println("Event channel was closed. Terminating event listener.");
					return;
				}
				//
				// Fail over
				//
				if event.GetKind() == EK_RESUBMIT {
					mail, ok := event.GetPayload().(model.Email);
					if ok {
						result.outgoing <- mail;
					}
				}

				//
				// Remove dead service from schedule
				//
				if event.GetKind() == EK_FATAL {
					provider,ok := event.GetPayload().(MTAProvider);
					if ok {
						result.scheduler.RemoveProviderFromService(provider);
						if len(result.scheduler.GetProviders()) < 1 {
							result.log.Println("No MTA Providers left we shutdown.")
							result.Stop();
							result.events <- NewEvent(EK_GOOD_BYE,errors.New("Container shuts down"));
							return;
						}
					}
				}

				result.events <- event;
			}
		}();
	}

	// upon receiving an e-mail to send, schedule and dispatch
	go func() {
		for {

			var email = <-result.outgoing;
			provider := result.scheduler.Schedule();
			if provider == nil {
				result.log.Println("Scheduler gave nil Provider");
				n := len(result.outgoing);
				for j := 0; j < n; j++ {
					result.events <- NewEvent(EK_INFORM_USER,errors.New("No MTAs availble email not sent: "+
						email.GetHeader(model.EML_HDR_SUBJECT)));
				}
				return; // no more MTAs
			}
			result.log.Println("Scheduling mail for sending on " + provider.GetName());
			provider.GetOutgoing() <- email;
		}
	}();

	go result.forwardIncomingMailToBackend();
	go result.listenForSendBackEnd();

	return result;
}

// ---------------------------------------------------------------
//
//
// Listening for ClientAPI to request sending e-mail
//
//
// ---------------------------------------------------------------
func (c *DefaultMTAContainer) listenForSendBackEnd() {

	var mux = http.NewServeMux();
	mux.HandleFunc("/sendmail",
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close();


			wireMail := model.NewWireEMailFromReader(r.Body);
			if wireMail == nil {
				log.Println("Error: Could not deserialise data.");
				http.Error(w, "Unable to deserialize request. Revise your email.", http.StatusBadRequest);
				return;
			}


			// For multiple recipients we send one email object each to
			// get more fair scheduling of providers.
			tos := strings.Split(wireMail.To(), ",");
			for m := range tos {
				wireMail.SetTo(tos[m]);
				c.GetOutgoing() <- wireMail.ToEmail();
			}

		});

	http.ListenAndServe(utilities.MTA_LISTENS_FOR_SEND_BACKEND, mux);

}


//
// Someday we may be able to receive mails too. mail.bitlab.dk is configured
// s.t. mails can be delivered via MailGun and Amazon. We would need to implement an
// open port in the providers implementing Web-API delivery of email...
//
// When that is done we would forward such inbound e-mails to the BackEnd for storage
// for later retrieval by the ClientAPI on behalf of an Inbox-display UI component.
//
//
func (c *DefaultMTAContainer) forwardToBackend(mail model.Email) {

	c.log.Println("Received email for delivery, forwarding it to Receive Backend");


	serJemail,errSerJemail := json.Marshal(mail);
	if errSerJemail != nil {
		println("Cannot serialise message.");
		return;
	}

	var url = "http://localhost"+utilities.RECEIVE_BACKEND_LISTENS_FOR_MTA+"/newmail";
	log.Println("Contacting receive back end at: "+url);


	resp, err := http.Post(url,"text/json",bytes.NewReader(serJemail));
	defer resp.Body.Close();
	if err != nil {
		c.events <- NewEvent(EK_WARNING,err);
	}
}



func (c * DefaultMTAContainer) forwardIncomingMailToBackend() {

	for {
		select {
		case receivedMail := <-c.GetIncoming():
			c.forwardToBackend(receivedMail);
		case cmd := <-c.cmd:
			if cmd == commandprotocol.CMD_MTA_PROVIDER_SHUTDOWN {
				c.log.Println("Forwading of incoming mails has been discontinued by shutdown command");
				return;
			}
		}
	}
}

