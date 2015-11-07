package main
import (
	"mail.bitlab.dk/model"
	"net/http"
	"mail.bitlab.dk/utilities"
	"fmt"
	"encoding/json"
	"bytes"
	"log"
)


func forwardToReceiver(mail model.Email) {

	log.Println("Received email for delivery, forwarding it to Receive Backend");

	jemail := model.EmailImpl{ Headers: mail.GetHeaders(),
									   Content: mail.GetContent()};

	serJemail,errSerJemail := json.Marshal(jemail);
	if errSerJemail != nil {
		println("Cannot serialise message.");
		return;
	}

	var url = "http://localhost"+utilities.RECEIVE_BACKEND_LISTENS_FOR_MTA+"/newmail";
	log.Println("Contacting receive back end at: "+url);
	response,err := http.Post(url,"text/json", bytes.NewReader(serJemail));

	if err != nil {
		configurationError("Could not connect to receive back-end."+err.Error());
	} else {
		fmt.Println(response);
	}

}

