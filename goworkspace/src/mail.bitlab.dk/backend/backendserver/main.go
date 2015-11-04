package main
import (
	"mail.bitlab.dk/utilities"
	"os"
	"mail.bitlab.dk/backend"
	"bufio"
	"log"
)



func main() {

	var i_am = utilities.PrintGreeting(os.Stdout);

	var store = backend.NewMemoryStore();
	var receiver = backend.NewReceiveBackend(store);
	var sender = backend.NewSendBackend(store);

	for {
		println("type q to quit.");
		var in = bufio.NewReader(os.Stdin);
		var input,_,err = in.ReadLine();
		if err != nil {
			log.Println("Could not read line from Stdin... leaving");
			receiver.Stop();
			sender.Stop();
			return;
		}

		if (input == "q") {
			receiver.Stop();
			sender.Stop();
			return;
		}
	}
}