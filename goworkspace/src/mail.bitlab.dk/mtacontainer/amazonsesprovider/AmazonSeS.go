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
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"mail.bitlab.dk/utilities"
	"github.com/aws/aws-sdk-go/aws/awserr"
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
	amazonApi       *ses.SES;
	log             *log.Logger;
	failureStrategy mtacontainer.FailureStrategy;
}


const (
	AWS_CNF_PASSPHRASE = "passphrase";
	AWS_CNF_API_KEY_ID = "keyid";
	AWS_CNF_ENC_SECRET_KEY = "secretkey";
	AWS_CNF_SECRET_KEY_LEN = "secretkeylen";
)

// ------------------------------------------------------
//
// MTA Provider API
//
// ------------------------------------------------------
func (ths *AmazonMtaProvider) Stop() {
	ths.command <- commandprotocol.CMD_MTA_PROVIDER_SHUTDOWN;
	close(ths.incoming);
	close(ths.outgoing);
	close(ths.command);
	close(ths.events);

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

func New(log *log.Logger, config map[string]string, fs mtacontainer.FailureStrategy) mtacontainer.MTAProvider {
	var result = new(AmazonMtaProvider);
	result.log = log;
	result.incoming = make(chan model.Email);
	result.outgoing = make(chan model.Email);
	result.events = make(chan mtacontainer.Event);
	result.command = make(chan commandprotocol.Command);
	result.failureStrategy = fs;
	awsLogger := aws.NewDefaultLogger();
	myCredentials := credentials.NewStaticCredentials(config[AWS_CNF_API_KEY_ID],
		utilities.DecryptApiKey(config[AWS_CNF_PASSPHRASE],
			config[AWS_CNF_ENC_SECRET_KEY],
			goh.StrToInt(config[AWS_CNF_SECRET_KEY_LEN])), "");
	mySession := session.New(&aws.Config{Region: aws.String("us-west-2"), Credentials: myCredentials, Logger: awsLogger});
	result.amazonApi = ses.New(mySession);
	if result.amazonApi == nil {
		return nil;
	}
	go result.serviceSendingEmails();
	log.Println(result.GetName() + " MTA Going up")
	return result;
}


// ------------------------------------------------------------
//
// We use the Amazon SeS function SendEmail found here
//
//
// ------------------------------------------------------------
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

		case mail := <-ths.outgoing:


			params := &ses.SendRawEmailInput{RawMessage: &ses.RawMessage {
				Data: mail.GetRaw(),
			} };


			//
			// Amazon errors are described here:
			// http://docs.aws.amazon.com/ses/latest/APIReference/CommonErrors.html
			// we take the interpretation that 500 and above means the service is suffering
			// down time. 400 means the message is somehow not properly formatted e.g.
			// it doesn't conform to MIME-rules and no further processing of the message
			// will take place.
			//
			resp, err := ths.amazonApi.SendRawEmail(params)



			if err != nil {
				var ok bool = false;
				var amazonError awserr.RequestFailure;
				amazonError, ok = err.(awserr.RequestFailure);

				//
				// is it a request error we can determine what actions to take based on the
				// status code.
				// Otherwise we are a bit in the dark assuming the worst we invoke the failureStrategy.
				//
				//
				if ok {
					code := amazonError.StatusCode();

					// service suffers resubmit and invoke failureStrategy
					if code > 500 {
						ths.events <- mtacontainer.NewEvent(mtacontainer.EK_RESUBMIT, err, mail);
						if (ths.failureStrategy.Failure(mtacontainer.EK_CRITICAL) == true) {
							// Report Amazon SES as down, Ask the MTA to fail over this service.
							//ths.events <- mtacontainer.NewEvent(mtacontainer.EK_FATAL, err, mailToSend);
							ths.log.Println("Amazon Service Provider had too many errors and shuts down.");
							ths.Stop();
						}
						continue;
					}

					// message is malformed, discard it but inform the log
					ths.events <- mtacontainer.NewEvent(mtacontainer.EK_INFORM_USER, errors.New("Message not sent: " + err.Error()), mail.GetSessionId());
					continue;
				} else {

					// A generic Amazon error occurred.
					if (ths.failureStrategy.Failure(mtacontainer.EK_CRITICAL) == true) {
						// Report Amazon SES as down, Ask the MTA to fail over this service.
						//ths.events <- mtacontainer.NewEvent(mtacontainer.EK_FATAL, err, mailToSend);
						ths.log.Println("Amazon Service Provider had too many errors and shuts down.");
						ths.Stop();
						continue;
					}
				}
			}

			log.Println(resp);
			ths.events <- mtacontainer.NewEvent(mtacontainer.EK_INFORM_USER,
				errors.New("Mail Delivered With Amazon SeS Successfully"),mail.GetSessionId());
			ths.failureStrategy.Success();
		}

	}
}


func (ths * AmazonMtaProvider) serviceReceivingEmails() {
	// TODO(rwz): Implement listening for incoming e-mail delivered by Amazon SeS.
}