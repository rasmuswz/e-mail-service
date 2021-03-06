//
//
// The ClientApi entry point.
//
// The process running this program will serve browser http requests on https://<domain>:<port>/*
// for a <web root> directory provided on up start.
//
// The special paths http://<domain>:<port>/go.api/*  will require
// BasicAuth in the http-header and (if valid) carry out ClientApi commands.
// The go.api/* paths are intended to serve the AJAX-requests
// issued by our JS-application running in the browser.
//
// See the {clientapi} package for more information.
//
// Author: Rasmus Winther Zakarias
//
package main

import (
	"mail.bitlab.dk/clientapi";
	"os"
	"mail.bitlab.dk/mtacontainer"
	"log"
	"strconv"
	"mail.bitlab.dk/utilities"
)


func main() {

	var i_am = utilities.PrintGreeting(os.Stdout);
	var accepting_args = " <web root> <port>";
	if (len(os.Args) < 3) {
		println(i_am+accepting_args);
		os.Exit(-1);
	}


	var root = os.Args[1];
	port,portErr := strconv.ParseInt(os.Args[2],0,16);

	if portErr != nil {
		println(i_am+" not happy with port argument "+os.Args[2]+" should be a 16-bit in [0;65535].");
		os.Exit(-2);
	}

	var clientApi = clientapi.New(root,int(port));
	var events = clientApi.GetEvent();
	var curDir,_ = os.Getwd();
	println(i_am+" Client API serving "+root+" at :"+strconv.Itoa(int(port))+" from directory "+curDir);

	//
	// The ClientApi server implements the {Health} interface.
	// For now we simple listens for events and prints these to the log.
	// The EK_DOWN event makes us shutdown.
	//
	for {

		select {
		case e := <-events:
			if e.GetKind() == mtacontainer.EK_FATAL {
				return;
			} else {
				log.Println(e.GetError().Error());
			}
		}
	}
	println(i_am+" leaving, bye");
}