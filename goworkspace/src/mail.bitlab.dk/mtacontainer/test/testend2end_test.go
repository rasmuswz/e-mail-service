package test
import (
	"testing"
	"mail.bitlab.dk/mtacontainer/loopbackprovider"
	"mail.bitlab.dk/utilities"
	"mail.bitlab.dk/mtacontainer"

)


// ------------------------------------------
//
//
// End 2 End test of the MTAContainer with Failure strategy
// and RoundRobinSchedule.
//
//
func TestEnd2End(t *testing.T) {


	var provider1 = loopbackprovider.New(utilities.GetLogger("loop1"),mtacontainer.NewThressHoldFailureStrategy(12));
	var provider2 = loopbackprovider.New(utilities.GetLogger("loop2"),mtacontainer.NewThressHoldFailureStrategy(12));

	scheduler := mtacontainer.NewRoundRobinScheduler([]mtacontainer.MTAProvider{provider1,provider2});

	container := mtacontainer.New(scheduler);
	c := make(chan int);
	go func() {
		for {
			select {
			case <-container.GetEvent():
			case <-c:
				return;
			}
		}
	}();

	mail1 := FreshTestMail(provider1,"rwl@cs.au.dk");
	mail2 := FreshTestMail(provider2,"rwl@cs.au.dk");


	container.GetOutgoing() <- mail1;
	mail1prime :=<-provider1.GetIncoming();
	container.GetOutgoing() <- mail2;
	mail2prime := <-provider2.GetIncoming();


	if mail1prime != mail1 {
		t.Error("Expected to see object mail1, buy saw different object");
	}

	if mail2prime != mail2 {
		t.Error("Expected to see object mail1, buy saw different object");
	}

	c<-1;

}