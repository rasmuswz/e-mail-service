package main
import (
	"mail.bitlab.dk/mtacontainer"
	"mail.bitlab.dk/utilities"
	"mail.bitlab.dk/mtacontainer/mailgun"
	"mail.bitlab.dk/mtacontainer/amazonsesprovider"
	"mail.bitlab.dk/mtacontainer/mandrill"
	"mail.bitlab.dk/mtacontainer/sendgridprovider"
	"mail.bitlab.dk/mtacontainer/smtpprovider"
	"mail.bitlab.dk/mtacontainer/loopbackprovider"
	"os"
	"fmt"
	"bufio"
	"log"
)


func GetLoopBackContainer() (mtacontainer.MTAContainer, mtacontainer.Scheduler) {
	var provider = loopbackprovider.New(utilities.GetLogger("LoopBack"));
	var scheduler= mtacontainer.NewRoundRobinScheduler([]mtacontainer.MTAProvider{provider});
	return mtacontainer.New(scheduler), scheduler;
}

func GetPassphraseFromArgOrTerminal() string {
	var passphrase []byte = nil;

	if (len(os.Args) < 2) {
		var passphraseErr error;
		var _ bool;
		fmt.Println("Parts of the source code contains API keys which are encrypted");
		fmt.Println("Please provide the password: ");
		passphrase, _, passphraseErr = bufio.NewReader(os.Stdin).ReadLine();
		if passphraseErr != nil {
			log.Fatalln("Could not read user input.");
		}
	} else {
		return os.Args[1];
	}

	return string(passphrase);
}

func GetProductionMTAContainer() (mtacontainer.MTAContainer,mtacontainer.Scheduler) {
	var passphrase = GetPassphraseFromArgOrTerminal();
	var mailGunConfig = mailgunprovider.BitLabConfig(passphrase);
	providers := make([]mtacontainer.MTAProvider, 6);
	providers[0] = mailgunprovider.New(utilities.GetLogger("MailGun"), mailGunConfig);
	providers[1] = amazonsesprovider.New(utilities.GetLogger("amazonSES"), mtacontainer.NewThressHoldFailureStrategy(2));
	providers[2] = mandrill.New(utilities.GetLogger("Mandill"));
	providers[3] = sendgridprovider.New(utilities.GetLogger("SendGrid"));
	providers[4] = smtpprovider.New(utilities.GetLogger("Smtp [localhost:587]"), "localhost", 587);
	providers[5] = smtpprovider.New(utilities.GetLogger("Smtp [localhost:25]"), "localhost", 25);
	var scheduler mtacontainer.Scheduler = mtacontainer.NewRoundRobinScheduler(providers);
	return mtacontainer.New(scheduler),scheduler;
}