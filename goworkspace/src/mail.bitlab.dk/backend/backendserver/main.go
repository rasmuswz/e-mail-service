package main
import (
	"mail.bitlab.dk/utilities"
	"os"
	"mail.bitlab.dk/backend"
	"bufio"
	"log"
	"strings"
)



func main() {

	utilities.PrintGreeting(os.Stdout);

	var store = backend.NewMemoryStore();
	var receiver = backend.NewReceiveBackend(store);
	var sender = backend.NewSendBackend(store);
	go func() {

		select {
		case e := <-receiver.GetEvent():
			println(e.GetError().Error());

		}

	}();


	for {
		println("type q to quit.");
		var in = bufio.NewReader(os.Stdin);
		var input, _, err = in.ReadLine();
		if err != nil {
			log.Println("Could not read line from Stdin... leaving");
			receiver.Stop();
			sender.Stop();
			return;
		}

		if (strings.Compare("q", string(input)) == 0) {

			receiver.Stop();
			sender.Stop();
			return;
		}

		println("Executing command \"" + string(input) + "\"");
	}
}