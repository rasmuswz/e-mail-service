/**
 *
 * Main entry point for the e-mail service.
 *
 * This program will spawn the full service stack:
 * - MTAContainer
 * - Backends
 * - JSon Store
 * - ClientAPI
 *
 * Author: Rasmus Winther Zakarias
 *
 */
package main

import (
	"fmt"
	"mail.bitlab.dk/mtacontainer"
	"log"
	"os"
	"bufio"
	"mail.bitlab.dk/mtacontainer/mailgun"
)


const GREETING = "Uber Challenge - GeoMail by Bitlab - The localized e-mail"
const COPYRIGHT= "All rights are reserved (C) Rasmus Winter Zakarias"

func main()  {

	fmt.Println(GREETING);
	fmt.Println(COPYRIGHT+"\n\n");

	// get Stdout logger
	log := log.New(os.Stdout,"[Log]",log.Lshortfile | log.Ltime | log.Ldate);
	log.Print("Initial logger created, hi Log ");

	// fire up the MTA container
	mp  := mailgunprovider.NewMailGun(log);
	providers := make([]mtacontainer.MTAProvider,1);
	providers[0] = mp;
	container := mtacontainer.CreateMTAContainer(mtacontainer.NewRoundRobinScheduler(providers));

	go func() {
		for {
			var e,ok = <-container.GetEvent();

			if (!ok) {
				return;
			}

			log.Println(e.GetError().Error());
		}
	}();

	in := bufio.NewReader(os.Stdin);
	input := "";
	for input != "q" {
		i,_,err := in.ReadLine();
		if err != nil {
			panic(err);
		}
		input = string(i);
	}
	return ;
}