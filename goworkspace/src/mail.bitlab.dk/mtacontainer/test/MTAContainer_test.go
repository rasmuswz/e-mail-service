//
//
// In this file the MTA Container is tested.
//
// Moved to utilities/tests to resolve cyclic dependency
//

package test
import (
"testing"
	"mail.bitlab.dk/mtacontainer"
)



func Test_create(t *testing.T) {

	schedule := NewSchedulerMock();

	container := mtacontainer.New(schedule);


	if container == nil {
		t.Error("Container should not be nil");
	}
}

// ------------------------------------
//
// UUT: MTAContainer
//
// ------------------------------------
func Testemail_outgoing(t *testing.T) {

	provider := NewMockMTAProvider(t);

	schedule := NewSchedulerMock();

	container := mtacontainer.New(schedule);

	mail := FreshTestMail(provider,"rwl@cs.ua.dk");

	container.GetOutgoing() <- mail;
}