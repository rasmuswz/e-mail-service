package utilities
import (
	"os"
	"io"
	"fmt"
	"strings"
	"log"
)


const GREETING = "Uber Challenge - BitMail by Bitlab - The reliable sender"
const COPYRIGHT= "All rights are reserved (C) Rasmus Winther Zakarias"


func PrintGreeting(s io.Writer) string {
	var i_am = os.Args[0];
	if strings.Contains(i_am,"/") {
		i_am = i_am[strings.LastIndex(i_am, "/") + 1:];
	}
	fmt.Fprintln(s,GREETING);
	fmt.Fprintln(s,COPYRIGHT+" - "+i_am+" going up.");
	return i_am;
}

func GetLogger(componentName string, targets ...io.Writer) *log.Logger{

	if (len(targets) > 1) {
		println("GetLogger, multiple targets not supported, yet.");

		// we could create an aggregate logger that takes several
		// Writers and posts message to all of them.
	}

	if (len(targets) == 0) {
		targets = append(targets,os.Stdout);
	}

	return 	log.New(targets[0] ,"["+componentName+"]",log.Lshortfile | log.Ltime | log.Ldate);

}