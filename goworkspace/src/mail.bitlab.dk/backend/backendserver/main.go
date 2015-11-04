package main
import (
	"mail.bitlab.dk/utilities"
	"os"
)



func main() {

	var i_am = utilities.PrintGreeting(os.Stdout);

	var mtaContainerAttachUrl string = "";
	var mtaContainerSendUrl string = "";
	var clientApi string = "";

	var accepting_args = " <mta container attach url> <mta container send url> <client api url>";

	if (len(os.Args) < 4) {
		println(i_am+accepting_args);
		os.Exit(-1);
	}

	mtaContainerAttachUrl = os.Args[1];
	mtaContainerSendUrl = os.Args[2];
	clientApi = os.Args[3];

	println("TODO(rwz): "+i_am+ " not implemented, implement: ");

	println("- We need to start JSon Storage ");

	println("- The Receiver for stuffing e-mails into Storage getting notified from: "+mtaContainerAttachUrl);

	println("- The sender for throwing e-mails at the MTA container found at "+mtaContainerSendUrl);

	println("- Accept commands from ClientApi at: "+clientApi);

}