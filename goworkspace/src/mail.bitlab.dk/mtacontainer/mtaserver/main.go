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

	"io/ioutil"
	"strings"
	"encoding/base64"
)



func forwardIncomingMailToReceiver(container mtacontainer.MTAContainer) {

	for {
		select {
		case receivedMail := <-container.GetIncoming():
			forwardToReceiver(receivedMail);
		}
	}
}

func listenForSendBackEnd(container mtacontainer.MTAContainer) {
	type serEmail struct {
		Headers map[string]string;
		Content string;
	}

	var mux = http.NewServeMux();
	mux.HandleFunc("/sendmail",
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close();


			data, dataErr := ioutil.ReadAll(r.Body);
			if dataErr != nil {
				log.Println("Error: Could not deserialise data.");
				http.Error(w, dataErr.Error(), http.StatusInternalServerError);
				return;
			}

			var jemail serEmail = serEmail{};

			err := json.Unmarshal(data, &jemail);
			if err != nil {
				log.Println("Error:\n" + err.Error() + " Could not deserialise email data:\n" + string(data));
				http.Error(w, "Error: Could not deserialise email data", http.StatusInternalServerError);
				return;
			}

			//
			// Content-Type: text/plain; charset=ISO-8859-1
			// Content-transfer-encoding: base64
			//

			jemail.Headers["Content-Type"] = "text/plain; charset=UTF-8";
			jemail.Headers["Content-transfer-encoding"] = "base64";
			jemail.Content = base64.StdEncoding.EncodeToString([]byte(jemail.Content));

			tos := strings.Split(jemail.Headers["To"], ",");
			for m := range tos {
				jemail.Headers["To"] = strings.Trim(tos[m]," ");
				container.GetOutgoing() <- model.NewMailS(jemail.Content, jemail.Headers);
			}

		});

	http.ListenAndServe(utilities.MTA_LISTENS_FOR_SEND_BACKEND, mux);

}

func promptUserToShutDownService() {
	//
	// Keep the main thread alive until the uses presse "q"
	//
	for {
		println("type q to quit.");
		var in = bufio.NewReader(os.Stdin);
		var input, _, err = in.ReadLine();
		if err != nil {
			log.Println("Could not read line from Stdin... leaving");
			return;
		}

		if (strings.Compare("q", string(input)) == 0) {

			return;
		}

		println("Executing command \"" + string(input) + "\"");
	}
}


func configurationError(reason string) {

	println("CONFIGURATION ERROR");
	println("The System cannot continue for the following reason: ");
	println(reason);
	println("Please fix this problem and re-execute " + os.Args[0]);

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
	var container, _ = GetProductionMTAContainer();
	//var container, scheduler = GetLoopBackContainer();

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
	for {
		e := <-container.GetEvent();
		log.Println(e.GetError());
	}

	return;
}
