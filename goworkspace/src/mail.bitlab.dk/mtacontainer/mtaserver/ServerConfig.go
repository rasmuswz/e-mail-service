package main
import (
	"mail.bitlab.dk/mtacontainer"
	"mail.bitlab.dk/utilities"
	"mail.bitlab.dk/mtacontainer/mailgunprovider"
	"mail.bitlab.dk/mtacontainer/amazonsesprovider"
	"mail.bitlab.dk/mtacontainer/sendgridprovider"
	"mail.bitlab.dk/mtacontainer/loopbackprovider"
	"os"
	"fmt"
	"bufio"
	"log"
)

const ERROR_THRESHOLD = 2
func GetLoopBackContainer() (mtacontainer.MTAContainer, mtacontainer.Scheduler) {
	var provider = loopbackprovider.New(utilities.GetLogger("LoopBack"),mtacontainer.NewThressHoldFailureStrategy(ERROR_THRESHOLD));
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
	var amazonConfig = amazonsesprovider.BitLabConfig(passphrase);
	var sendgridConfig = sendgridprovider.BitLabConfig(passphrase);
	providers := make([]mtacontainer.MTAProvider, 5);
	providers[0] = mailgunprovider.New(utilities.GetLogger("MailGun"), mailGunConfig,mtacontainer.NewThressHoldFailureStrategy(ERROR_THRESHOLD));
	providers[1] = amazonsesprovider.New(utilities.GetLogger("amazonSES"), amazonConfig, mtacontainer.NewThressHoldFailureStrategy(ERROR_THRESHOLD));
	providers[2] = sendgridprovider.New(utilities.GetLogger("SendGrid"), sendgridConfig, mtacontainer.NewThressHoldFailureStrategy(ERROR_THRESHOLD));
	var scheduler mtacontainer.Scheduler = mtacontainer.NewRoundRobinScheduler(providers);
	return mtacontainer.New(scheduler),scheduler;
}