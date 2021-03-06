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
	"mail.bitlab.dk/utilities"
	"net/http"
	"strings"
	"log"
)

func postUserMessageWithClientAPI(message, sessionId string) {
	_,e := http.Post("http://localhost"+utilities.CLIENT_API_LISTEN_FOR_MTA+"/mta/usermessage?SessionId="+sessionId,"text/plain",strings.NewReader(message));
	if e != nil {
		log.Println("[MTAServer] Failed to poste user message:\n"+message+"\nwith ClientAPI:\n"+e.Error());
	}
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
	var container, _ = GetProductionMTAContainer();
	//var container, scheduler = GetLoopBackContainer();

	//
	// Print Events that occurs until GOOD_BYE.
	//
	for {
		e := <-container.GetEvent();
		log.Println(e.GetError());
		if e.GetKind() == mtacontainer.EK_GOOD_BYE {
			break;
		}

		if e.GetKind() == mtacontainer.EK_INFORM_USER {
			postUserMessageWithClientAPI(e.GetError().Error(),e.GetPayload().(string));
		}
	}

	return;
}
