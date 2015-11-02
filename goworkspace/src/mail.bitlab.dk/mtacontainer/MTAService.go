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
)

type EventKind uint32;
const (
	// no problem
	EK_OK = 0x00;
	// reporting entity is out of service
	EK_DOWN = 0x01;
	// reporting entity suffered a time out
	// but we haven't given up yet will try again
	EK_TIMEOUT = 0x02;
	// reporting entity beats (it is up and well)
	EK_BEAT = 0x03;
)

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
	err error;
	kind EventKind;
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

func NewEvent(kind EventKind,e error) Event {
	var r *defaultEvent = new(defaultEvent);
	r.kind = kind;
	r.err = e;
	r.time = time.Now();
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
	incoming chan model.Email;
	outgoing chan model.Email;
	events chan Event;
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
// Public constructor for creating an MTAContainer.
//
//
func CreateMTAContainer(scheduler Scheduler) MTAContainer {
	if (scheduler == nil) {
		return nil;
	}

	var result = new(DefaultMTAContainer);
	result.providers = scheduler.getProviders();
	result.incoming = make(chan model.Email);
	result.outgoing = make(chan model.Email);
	result.events   = make(chan Event);
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
		provider.GetOutgoing() <- email;
	}();

	return result;
}


// ------------------------------------------------------
//
// Round robin MTA Provider scheduler
//
// ------------------------------------------------------
type RoundRobinScheduler struct {
	current int;
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

func NewRoundRobinScheduler(providers []MTAProvider) Scheduler {
	if (len(providers) < 1) {
		return nil;
	}

	var rrs = new(RoundRobinScheduler);
	rrs.current = 0;
	rrs.providers = providers;
	return rrs;
}