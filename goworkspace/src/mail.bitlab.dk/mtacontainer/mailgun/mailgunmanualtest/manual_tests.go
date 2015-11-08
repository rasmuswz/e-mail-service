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
	"mail.bitlab.dk/utilities/tests"
	"mail.bitlab.dk/mtacontainer/mailgun"
)


func main() {
	mailgunMTAProvider := mailgunprovider.New(utilities.GetLogger("MailGun"),
		mailgunprovider.BitLabConfig("UberChallenge"),tests.NewMockFailureStrategy());
	tests.ManuallyVerifyEmailSend(mailgunMTAProvider);
}




