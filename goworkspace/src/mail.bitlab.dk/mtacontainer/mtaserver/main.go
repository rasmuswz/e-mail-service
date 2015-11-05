/**
 *
 * Main entry point for the mta-container service.
 *
 *
 * Author: Rasmus Winther Zakarias
 *
 */
package main

import (

	"mail.bitlab.dk/mtacontainer"
	"os"
	"bufio"
	"mail.bitlab.dk/utilities"
	"mail.bitlab.dk/model"
	"log"
	"net/http"
	"encoding/json"
)



func handleErrorFromMtaContainer(log *log.Logger,
								 container mtacontainer.MTAContainer,
								 scheduler mtacontainer.Scheduler) {
	for {
		var e, ok = <-container.GetEvent();

		if (!ok) {
			return;
		}

		//
		// Container requests e-mail to be submitted else where.
		//
		if e.GetKind() > mtacontainer.EK_RESUBMIT {
			var resubmitMail, ok = e.GetPayload().(model.Email); // syntax for typeof
			// we here as if payload is an model.Email object.
			if ok {
				container.GetOutgoing() <- resubmitMail;
			}
		}

		//
		// Container Died
		//
		if e.GetKind() == mtacontainer.EK_FATAL {
			var provider, ok = e.GetPayload().(mtacontainer.MTAProvider);
			if ok {
				log.Println("Provider \"" + provider.GetName() + " is permanently down and now decommissioned.");
				if scheduler.RemoveProviderFromService(provider) < 1 {
					container.Stop();
					log.Println("We have no services able of performing any server, please visit configuration ");
				}
			}
			log.Println("Something fatal happened to something not an MTAProvider :-(");
		}


		log.Println(e.GetError().Error());
	}

}

func forwardIncomingMailToReceiver(container mtacontainer.MTAContainer) {

	for {
		select {
		case receivedMail := <-container.GetIncoming():
			forwardToReceiver(receivedMail);
		}
	}
}

func listenForSendBackEnd(container mtacontainer.MTAContainer) {

	var mux = http.NewServeMux();
	mux.HandleFunc("/sendmail", func(w http.ResponseWriter, r *http.Request) {
		var jDec = json.NewDecoder(r.Body);
		var jemail = model.EmailFromJSon{}
		err := jDec.Decode(&jemail);
		if err != nil {
			println("Server Failed to deserialise stream: "+err.Error());
		} else {
			container.GetOutgoing() <- model.NewEmailFromJSon(&jemail);
		}
		r.Body.Close();
	});

	http.ListenAndServe(utilities.MTA_LISTENS_FOR_SEND_BACKEND,mux);

}

func promptUserToShutDownService() {
	println("MTAServer running type \"q\"<enter> to stop it.");
	in := bufio.NewReader(os.Stdin);
	input := "";
	for input != "q" {
		i, _, err := in.ReadLine();
		if err != nil {
			panic(err);
		}
		input = string(i);
	}
}


func configurationError(reason string) {

	println("CONFIGURATION ERROR");
	println("The System cannot continue for the following reason: ");
	println(reason);
	println("Please fix this problem and re-execute "+os.Args[0]);

	os.Exit(-1);

}

func main() {

	//
	// say hello
	//
	utilities.PrintGreeting(os.Stdout);


	//
	// Initialize logger for stdout for this mtaserver.
	//
	log := utilities.GetLogger("mtaserver");
	log.Print("Initial logger created, hi Log ");


	//
	// Start Container (ask for passphrase if necessary)
	//
	// var container,scheduler = GetProductionMTAContainer();
	var container,scheduler = GetLoopBackContainer();

	//
	// Error handling
	//
	go handleErrorFromMtaContainer(log,container,scheduler);

	//
	// Forward incoming e-mail to Receiver Back-End
	//
	go forwardIncomingMailToReceiver(container);

	//
	// Listen for outgoing e-mail from Send-End end
	//
	go listenForSendBackEnd(container);


	//
	// Take control for terminal
	//
	promptUserToShutDownService();
	return;
}