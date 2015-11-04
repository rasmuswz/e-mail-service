package utilities
import (
	"os"
	"io"
	"fmt"
	"strings"
)


const GREETING = "Uber Challenge - GeoMail by Bitlab - The localized e-mail"
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