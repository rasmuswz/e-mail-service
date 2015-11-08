// ------------------------------------------------------------
//
// We test here that the MTAContainer can fail-over
//
//
// Author: Rasmus Winther Zakarias
//
// ------------------------------------------------------------
package test
import (
	"mail.bitlab.dk/mtacontainer/loopbackprovider"
	"mail.bitlab.dk/utilities"
	"mail.bitlab.dk/mtacontainer"
	"testing"
	"log"
)

type FailImmediatelyFailureStrategy struct {

}

func (ths *FailImmediatelyFailureStrategy) Failure(f mtacontainer.EventKind) bool {
	return true;
}

func (ths *FailImmediatelyFailureStrategy) Success() {

}

func Test_all_failover(t *testing.T) {

	loop1 := loopbackprovider.New(utilities.GetLogger("loop1"), new(FailImmediatelyFailureStrategy));

	scheduler := mtacontainer.NewRoundRobinScheduler([]mtacontainer.MTAProvider{loop1});

	container := mtacontainer.New(scheduler);

	go func() {
		mail := FreshTestMail(loop1, "rwl@cs.au.dk");
		container.GetOutgoing() <- mail;
		for {
			select {
			case evt := <-container.GetEvent():
				log.Println(evt.GetError().Error());
			}
		}
	}();


}

func Test_one_failover(t *testing.T) {

	loop1 := loopbackprovider.New(utilities.GetLogger("loop1"), new(FailImmediatelyFailureStrategy));
	loop2 := loopbackprovider.New(utilities.GetLogger("loop2"), NewMockFailureStrategy());

	scheduler := mtacontainer.NewRoundRobinScheduler([]mtacontainer.MTAProvider{loop1, loop2});

	container := mtacontainer.New(scheduler);

	mail := FreshTestMail(loop1, "rwl@cs.au.dk");



	go func() {
		container.GetOutgoing() <- mail;
		for {
			select {
			case evt := <-container.GetEvent():
				log.Println(evt.GetError().Error());
			}
		}
	}();


	// the round robin scheduler will schedule {loop1} which fails immediately
	// we expect to see {mail} emerge on the {GetIncoming} channel of {loop2}
	// because it has been RESUBMITTET to {loop2} by the fail over mechanism in
	// the MTAContainer.



	mailPrime, ok := <-loop2.GetIncoming()

	if ok == false {
		t.Error("loop2 incoming failed to provide emai.");
	}


	if mailPrime != mail {
		t.Error("Wrong mail");
	}
}


