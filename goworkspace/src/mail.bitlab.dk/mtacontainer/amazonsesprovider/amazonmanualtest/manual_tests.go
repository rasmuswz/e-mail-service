//
// The Actual providers cannot be tested automatically entirely
//
// To check that an email actually arrives when providing an e-mail
// to the outgoing channel needs to be done manually. This we will
// do here.
//
// Author: Rasmus Winther Zakarias
//

package main
import (
	"mail.bitlab.dk/mtacontainer/amazonsesprovider"
	"mail.bitlab.dk/utilities"
	"mail.bitlab.dk/mtacontainer/test"
	"os"
)


func main() {

	if (len(os.Args) < 3) {
		println("manual_test <to address> <key>");
		return;
	}

	to := os.Args[1];
	key := os.Args[2];

	amazonMTAProvider := amazonsesprovider.New(utilities.GetLogger("Amazon"),
		amazonsesprovider.BitLabConfig(key), test.NewMockFailureStrategy());
	test.ManuallyVerifyEmailSend(amazonMTAProvider,to);
}




