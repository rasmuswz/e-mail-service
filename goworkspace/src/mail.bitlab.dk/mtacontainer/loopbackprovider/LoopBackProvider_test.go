package loopbackprovider_test


import (
	"testing"
	"mail.bitlab.dk/mtacontainer/loopbackprovider"
	"mail.bitlab.dk/utilities"
	"mail.bitlab.dk/mtacontainer"
	"mail.bitlab.dk/mtacontainer/test"
)


func TestLoopBack(t *testing.T) {

	loop1 := loopbackprovider.New(utilities.GetLogger("loop1"), mtacontainer.NewThressHoldFailureStrategy(1));

	mail := test.FreshTestMail(loop1, "test");

	c := make(chan int);

	go func() {
		for {
			select {
			case <-loop1.GetEvent():
			case <-c:
			}
		}
	}();

	loop1.GetOutgoing() <- mail;


	<-loop1.GetIncoming();
}