package test
import (
	"testing"

	"mail.bitlab.dk/mtacontainer"
)

func TestRoundRobinScheduler_Test(t *testing.T) {

	var mocks = []mtacontainer.MTAProvider{
		NewMockMTAProvider(t),
		NewMockMTAProvider(t),
		NewMockMTAProvider(t),
	};

	rrs := mtacontainer.NewRoundRobinScheduler(mocks);

	for i:=0; i < 400; i += 1 {
		if rrs.Schedule() != mocks[i%3] {
			t.Error("Wrong schedule.")
		}
	}



}