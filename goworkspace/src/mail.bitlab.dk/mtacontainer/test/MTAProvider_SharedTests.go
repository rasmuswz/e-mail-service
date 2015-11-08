//
// This test verifies black-box that a MTAproviders work by submitting
// an email for sending and letting the user verify that an email indeed
// arrived in he(r) Inbox.
//
// The test here is parameterized by the kind of MTAProvider
//
// Author: Rasmus Winther Zakarias
//


package test
import (
	"mail.bitlab.dk/mtacontainer"
	"mail.bitlab.dk/utilities/go"
	"strings"
)

// ---------------------------------------------------------
//
// We Check that the provider can send e-mail by manual
// verification.
//
// There is a main function for each provider invoking this
// function with an instance of the provider.
//
// See amazonsesprovider/amazonmanualtest/manualtest
//     mailgunprovider/mailgunmanualtest/manualtest
//     sendgridprovider/sendgridmanualtest/manualtest
//
//
//
// ---------------------------------------------------------

func ManuallyVerifyEmailSend(provider mtacontainer.MTAProvider, to string)  {

	//
	// Output all events from MTA Provider
	//
	c := make(chan int);
	go func() {
			for {
				select {
				case <- c:
						return;
				case e :=<-provider.GetEvent():
					println(e.GetError().Error());
				}
			}
	}();

	// Send verification email
	mail := FreshTestMail(provider, to);
	provider.GetOutgoing() <- mail;

	// wait for manual verification.
	line := goh.ReadLine("Did your receive an e-mail with subject \""+mail.GetHeaders()["Subject"][0]+
	"\" at "+ to +" [y/n]?");

	ok := strings.Compare("y",line) == 0;

	if ok == false {
		println("Manually Verify Email Send Failed.")
	} else {
		println("Yes mail successfully sent :-)")
	}

}




