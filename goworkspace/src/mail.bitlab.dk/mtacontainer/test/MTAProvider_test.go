//
// Unit Under Test: Any MTA Provider we test it black box here
//
//
// Author: Rasmus Winther Zakarias
//
package test
import (
	"testing"

	"mail.bitlab.dk/mtacontainer"
	"mail.bitlab.dk/mtacontainer/loopbackprovider"
	"mail.bitlab.dk/utilities"
)

// Testing that the MTA actually can send email is a manual task deferred to MTAProvider_SharedTests.go
// and each of the manual_tests.go files.

//
// An MTA Provider is expected to report EK_FATAL
// right before it shuts down.
//
//  Unit under test: Any MTAProvider
//
//
func Test_check_fatal_event_on_shutdown(t *testing.T) {
	c := make(chan int);
	events := make([]mtacontainer.Event, 0);
	provider := loopbackprovider.New(utilities.GetLogger("Loop"),NewMockFailureStrategy());

	go func() {
		for {
			select {
			case e := <-provider.GetEvent():
				events = append(events, e);
			case <-c:
				return;
			}
		}
	}();


	provider.Stop();


	if len(events) == 0 {
		t.Error("Expected at least one event EK_FATAL event...");
	}
	c <- 0;

	foundFatal := false;
	for i := range events {
		if events[i].GetKind() == mtacontainer.EK_FATAL {
			foundFatal = true;
		}
	}

	if foundFatal == false {
		t.Error("Expected one fatal event");
	}
}