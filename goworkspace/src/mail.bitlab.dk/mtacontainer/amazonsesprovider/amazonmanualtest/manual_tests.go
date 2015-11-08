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
	"mail.bitlab.dk/mtacontainer/amazonsesprovider"
	"mail.bitlab.dk/utilities"
	"mail.bitlab.dk/utilities/tests"
)


func main() {
	println("Encrypted Amazon Key: "+utilities.EncryptApiKey("5AKPej5pbSM2xaHgkG1Nzp5tPcRwztQ5Le8jqRsc","09de27"));
	println("Encrypted MailGunKey: "+utilities.EncryptApiKey(""));


	amazonMTAProvider := amazonsesprovider.New(utilities.GetLogger("Amazon"), tests.NewMockFailureStrategy());
	tests.ManuallyVerifyEmailSend(amazonMTAProvider);
}




