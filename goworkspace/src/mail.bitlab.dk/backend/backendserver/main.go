package main
import (
	"mail.bitlab.dk/utilities"
	"os"
	"mail.bitlab.dk/backend"
	"bufio"
	"log"
	"strings"
)


// ---------------------------------------------------------
//
// main enrty point for the backend server-application
//
// ---------------------------------------------------------
func main() {

	utilities.PrintGreeting(os.Stdout);

	var store = backend.NewMemoryStore();
	var receiver = backend.NewReceiveBackend(store);

	//
	// The backend implement our monitoring interface
	// HealthService which stipulates an Event channel
	// accessible by {GetEvent}. Below we make sure events
	// are written to stdout.
	//
	go func() {
		select {
		case e := <-receiver.GetEvent():
			println(e.GetError().Error());
		}
	}();


	//
	// Keep the main thread alive until the uses presse "q"
	//
	for {
		println("type q to quit.");
		var in = bufio.NewReader(os.Stdin);
		var input, _, err = in.ReadLine();
		if err != nil {
			log.Println("Could not read line from Stdin... leaving");
			receiver.Stop();
			return;
		}

		if (strings.Compare("q", string(input)) == 0) {
			receiver.Stop();
			return;
		}

		println("Executing command \"" + string(input) + "\"");
	}
}