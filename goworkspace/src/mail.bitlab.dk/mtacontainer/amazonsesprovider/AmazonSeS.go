package amazonsesprovider

import (
	"github.com/aws/aws-sdk-go/service/ses"
	"mail.bitlab.dk/model"
	"log"
	"mail.bitlab.dk/mtacontainer"
	"mail.bitlab.dk/utilities/commandprotocol"
	"mail.bitlab.dk/utilities/go"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"time"
);

//
// The Amazon Ses Mta provider
//
//
//MTAService;
// - GetIncoming() chan model.Email;
// - GetOutgoing() chan model.Email;
// - Stop();
//HealthService;
// - GetEvent() chan Event;
//GetName() string;
type AmazonMtaProvider struct {
	incoming        chan model.Email;
	outgoing        chan model.Email;
	events          chan mtacontainer.Event;
	command         chan commandprotocol.Command;
	amazonApi       ses.SES;
	log             *log.Logger;
	failureStrategy mtacontainer.FailureStrategy;
}


// ------------------------------------------------------
//
// MTA Provider API
//
// ------------------------------------------------------
func (ths *AmazonMtaProvider) Stop() {

}

func (ths *AmazonMtaProvider) GetOutgoing() chan model.Email {
	return ths.outgoing;
}

func (ths *AmazonMtaProvider) GetIncoming() chan model.Email {
	return ths.incoming;
}

func (ths * AmazonMtaProvider) GetEvent() chan mtacontainer.Event {
	return ths.events;
}

func (ths *AmazonMtaProvider) GetName() string {
	return "Amazon Simple Email Service provider";
}

// ---------------------------------------------------------
//
// Implementation
//
// ---------------------------------------------------------

func New(log *log.Logger, fs mtacontainer.FailureStrategy) mtacontainer.MTAProvider {
	var result = new(AmazonMtaProvider);
	result.log = log;
	result.incoming = make(chan model.Email);
	result.outgoing = make(chan model.Email);
	result.events = make(chan mtacontainer.Event);
	result.failureStrategy = fs;
	go result.serviceSendingEmails();
	log.Println(result.GetName() + " MTA Going up")

	return result;
}


//
//
//
func (ths *AmazonMtaProvider) serviceSendingEmails() {
	for {
		select {
		case cmd := <-ths.command:
		//
		// This chunk of code handles the SHUTDOWN command and other command
		//
			if cmd == commandprotocol.CMD_MTA_PROVIDER_SHUTDOWN {
				ths.log.Println(ths.GetName() + " is going down, received Shut Down command");
				ths.events <-
				mtacontainer.NewEvent(mtacontainer.EK_FATAL,
					errors.New(ths.GetName() + " received ShutDown command it is going down, bye."));
				return;
			} else {
				log.Println("Warning: Received command which has no defined action.");
				ths.events <- mtacontainer.NewEvent(mtacontainer.EK_WARNING,
					errors.New("I do not understand the command: " + goh.IntToStr(int(cmd))), ths);
			}

		case mailToSend := <-ths.outgoing:


		var headers map[string]string = mailToSend.GetHeaders();
		//
		// Handle outgoing email, e.g. forward it to AmazonSes
		//
			e := ses.SendEmailInput{
				Destination: &ses.Destination{
					ToAddresses: []*string{
						aws.String(headers[model.EML_HDR_TO]),
					},
				},
				Message: &ses.Message{
					Body: &ses.Body{
						Text: &ses.Content{
							Charset: aws.String("UTF-8"),
							Data: aws.String(mailToSend.GetContent()),
						},
					},
					Subject: &ses.Content{
						Data: aws.String(headers[model.EML_HDR_SUBJECT]),
						Charset: aws.String("UTF-8"),
					},
				},
				Source: aws.String(headers[model.EML_HDR_FROM]),
			};


			resp, err := ths.amazonApi.SendEmail(&e);
			if err != nil {
				ths.log.Println("[Critical] Failed to send E-mail. Please examine !!!",err.Error());

				if (ths.failureStrategy.Failure(mtacontainer.EK_CRITICAL) == false) {
					ths.events <- mtacontainer.NewEvent(mtacontainer.EK_WARNING, err, ths);
					// critical failure, fallback
					time.Sleep(time.Second * 2);
					// resubmit e-mail to the service
					ths.outgoing <- mailToSend;
				} else {
					// Report Amazon SES as down, Ask the MTA to fail over this service.
					ths.events <- mtacontainer.NewEvent(mtacontainer.EK_FATAL,err,mailToSend);
					ths.log.Println("Amazon Service Provider had too many errors and shuts down.");
					ths.Stop();
				}
				continue;
			}

			log.Println(resp);
			ths.failureStrategy.Success();
		}

	}
}


func (ths * AmazonMtaProvider) serviceReceivingEmails() {

}