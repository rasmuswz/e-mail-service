//
// The Actual providers cannot be tested automatically entirely
//
// To check that an email actually arrives when providing an e-mail
// to the outgoing channel needs to be done manually. This we wil;
// do here.
//
// Author: Rasmus Winther Zakarias
//

package main
import (
	"mail.bitlab.dk/utilities"
	"mail.bitlab.dk/mtacontainer/sendgridprovider"
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

	sendgridMTAProvider := sendgridprovider.New(utilities.GetLogger("SendGrid"),
		 sendgridprovider.BitLabConfig(key),test.NewMockFailureStrategy());
	test.ManuallyVerifyEmailSend(sendgridMTAProvider,to);
}




