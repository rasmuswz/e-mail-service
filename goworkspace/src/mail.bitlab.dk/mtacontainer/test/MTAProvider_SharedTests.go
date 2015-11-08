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
// The first command line argument shall be the email address
// we will send an e-mail to for verification.
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




