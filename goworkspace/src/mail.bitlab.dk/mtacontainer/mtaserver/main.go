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
)

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
	}

	return;
}
