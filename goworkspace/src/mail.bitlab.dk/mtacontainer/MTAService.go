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
	"net/http"
	"mail.bitlab.dk/utilities"
	"log"
	"encoding/json"
	"io/ioutil"
	"errors"
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
	EK_RESUBMIT = 0x07
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
	threshold int;
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

func NewThressHoldFailureStrategy(thresshold int) FailureStrategy{
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
	time time.Time;
	err  error;
	kind EventKind;
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
 * emails and a chanel for sending them.
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
 * A scheduler takes an array of MTAProviders and upon calling {schedule}
 * it returns the next MTAProvider to carry out an task.
 */
type Scheduler interface {
	schedule() MTAProvider;
	getProviders() []MTAProvider;
	RemoveProviderFromService(provider MTAProvider) int;
}



/*
 * The {MTAContainer} keeps track of a set of MTAProviders and provides an
 * unified interface for Email-service using a schedule for balancing the load
 * between the providers.
 *
 * Furthermore the MTAContainer provides a health service.
 */
type MTAContainer interface {
	getProviders() []MTAProvider;
	setScheduler(scheduler Scheduler);
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
}


func (d *DefaultMTAContainer) getProviders() []MTAProvider {
	return d.providers;
}

func (d *DefaultMTAContainer) setScheduler(scheduler Scheduler) {
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
	close(d.events);
}

//
// Public constructor for creating a DefaultMTAContainer.
//
func New(scheduler Scheduler) MTAContainer {
	if (scheduler == nil) {
		return nil;
	}

	var result = new(DefaultMTAContainer);
	result.providers = scheduler.getProviders();
	result.incoming = make(chan model.Email);
	result.outgoing = make(chan model.Email);
	result.events = make(chan Event);
	result.scheduler = scheduler;

	// Aggregate all provider incoming and event channels to those of the MTAContainer
	for p := range result.providers {
		var pp MTAProvider = result.providers[p];

		// incoming
		go func() {
			var in = pp.GetIncoming();
			var email = <-in;
			result.incoming <- email;
		}();

		// health events
		go func() {
			var evt = pp.GetEvent();
			var event, ok = <-evt;
			if (ok == false) {
				return;
			}
			result.events <- event;
		}();
	}

	// upon receiving an e-mail to send, schedule and dispatch
	go func() {
		var email = <-result.outgoing;
		provider := result.scheduler.schedule();
		log.Println("Scheduling mail for sending on "+provider.GetName());
		provider.GetOutgoing() <- email;
	}();

	return result;
}

func (ths *DefaultMTAContainer) receiveMailToBeSentFromSendBackEnd(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close();
	var mail model.EmailFromJSon = model.EmailFromJSon{};
	var data, dataErr = ioutil.ReadAll(r.Body);
	if dataErr != nil {
		ths.events <- NewEvent(EK_WARNING,dataErr,ths);
		http.Error(w,dataErr.Error(),http.StatusInternalServerError);
		return ;
	}

	mailErr := json.Unmarshal(data,&mail);
	if mailErr != nil {
		ths.events <- NewEvent(EK_WARNING,errors.New("Malformed email"),ths);
		http.Error(w,"Malformed email cannot send",http.StatusBadRequest);
		return;
	}

	var email = model.NewEmailFromJSon(&mail);

	ths.events <- NewEvent(EK_OK,errors.New("Forwarding Email to the MTAs."),ths);

	ths.GetOutgoing() <- email;

}

func (ths *DefaultMTAContainer) ListForSendBackEnd() {

	var mux = http.NewServeMux();
	mux.HandleFunc("/sendmail", ths.receiveMailToBeSentFromSendBackEnd);

	err := http.ListenAndServe(utilities.MTA_LISTENS_FOR_SEND_BACKEND, mux);
	if err != nil {
		log.Fatalln("Could not listen for Send Back End: " + err.Error());
	}

}

// ------------------------------------------------------
//
// Round robin MTA Provider scheduler
//
// ------------------------------------------------------
type RoundRobinScheduler struct {
	current   int;
	providers []MTAProvider;
}

func (rrs *RoundRobinScheduler) getProviders() []MTAProvider {
	return rrs.providers;
}

func (rrs *RoundRobinScheduler) schedule() MTAProvider {
	var result = rrs.providers[rrs.current];
	rrs.current = (rrs.current % len(rrs.providers));
	return result;
}

//
// If the given provider {mta} is Scheduled by this scheduler
// then remove it from the list of service providers.
//
func (rrs *RoundRobinScheduler) RemoveProviderFromService(mta MTAProvider) int {
	var found bool = false;
	for k := range rrs.providers {
		if mta == rrs.providers[k] {
			found = true;
		}
	}

	if found {
		var i int = 0;
		var newProviders = make([]MTAProvider, len(rrs.providers) - 1);
		for k := range rrs.providers {
			if rrs.providers[k] != mta {
				newProviders[i] = rrs.providers[k];
				i = i + 1;
			}
		}
		rrs.current = rrs.current % len(newProviders);
		rrs.providers = newProviders;
	}

	return len(rrs.providers);
}

func NewRoundRobinScheduler(providers []MTAProvider) Scheduler {
	if (len(providers) < 1) {
		return nil;
	}

	var rrs = new(RoundRobinScheduler);
	rrs.current = 0;
	rrs.providers = providers;
	return rrs;
}