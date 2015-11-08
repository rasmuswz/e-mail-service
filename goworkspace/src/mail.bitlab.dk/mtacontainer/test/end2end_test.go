package test
import (
	"testing"
//"mail.bitlab.dk/mtacontainer/loopbackprovider"
	"mail.bitlab.dk/utilities"
	"mail.bitlab.dk/mtacontainer"
	"log"
	"mail.bitlab.dk/mtacontainer/loopbackprovider"
)


// ------------------------------------------
//
//
// End 2 End test of the MTAContainer with Failure strategy
// and RoundRobinSchedule.
//
//
func TestEnd2End(t *testing.T) {


	//	mg := mailgunprovider.New(utilities.GetLogger("MG"),mailgunprovider.BitLabConfig("09fe27"),NewMockFailureStrategy());
	//	az := amazonsesprovider.New(utilities.GetLogger("MG"),amazonsesprovider.BitLabConfig("09fe27"),NewMockFailureStrategy());
	//  sg := sendgridprovider.New(utilities.GetLogger("MG"),sendgridprovider.BitLabConfig("09fe27"),NewMockFailureStrategy());
	var provider = loopbackprovider.New(utilities.GetLogger("loop1"), mtacontainer.NewThressHoldFailureStrategy(12));

	scheduler := mtacontainer.NewRoundRobinScheduler([]mtacontainer.MTAProvider{provider});

	container := mtacontainer.New(scheduler);

	mail1 := FreshTestMail(provider, "rwl@cs.au.dk");
	mail2 := FreshTestMail(provider, "rwl@cs.au.dk");
	mail3 := FreshTestMail(provider, "rwl@cs.au.dk");

	container.GetOutgoing() <- mail1;
	container.GetOutgoing() <- mail2;
	container.GetOutgoing() <- mail3;

	go func() {
		<-container.GetIncoming()
	}();

	i := 0;
	for {
		select {
		case e := <-container.GetEvent():
			log.Println("Reading event from container: " + e.GetError().Error());
			if i == 2 {
				return;
			}
			i = i + 1;
		}
	}
}