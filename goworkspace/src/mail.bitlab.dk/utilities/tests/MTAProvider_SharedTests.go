package tests
import (
	"mail.bitlab.dk/mtacontainer"
	"mail.bitlab.dk/utilities/go"
	"strings"
	"os"
)

// ---------------------------------------------------------
//
// We Check that the provider can send e-mail by manual
// verification.
//
// There is a main function for each provider invoking this
// function with an instance of the provider.
//
// The first command line argument shall be the email address
// we will send an e-mail to for verification.
//
// ---------------------------------------------------------

func ManuallyVerifyEmailSend(provider mtacontainer.MTAProvider)  {

	//
	// Get To address from args
	//
	if len(os.Args) < 3 {
		println("amazonmanualtest <email target for tests> <decryption key>");
		return;
	}
	to := os.Args[1];

	//
	// Check args
	//
	if provider == nil {
		println("Creating the provider failed.");
	}

	if len(strings.Split(to, "@")) != 2 {
		println("[ManualTest] Failed given first argument must be an email address.");
		return;
	}

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




