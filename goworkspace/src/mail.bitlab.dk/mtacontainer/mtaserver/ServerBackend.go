package main
import (
	"mail.bitlab.dk/model"
	"net/http"
	"mail.bitlab.dk/utilities"
	"fmt"
	"encoding/json"
	"bytes"
)


func forwardToReceiver(mail model.Email) {


	jemail := model.EmailFromJSon{ Headers: mail.GetHeaders(),
									   Content: mail.GetContent()};

	serJemail,errSerJemail := json.Marshal(jemail);
	if errSerJemail != nil {
		println("Cannot serialise message.");
		return;
	}

	response,err := http.Post(utilities.RECEIVE_BACKEND_LISTENS_FOR_MTA,
			                  "text/json", bytes.NewReader(serJemail));

	if err != nil {
		configurationError("Could not connect to receive back-end.");
	} else {
		fmt.Println(response);
	}

}

