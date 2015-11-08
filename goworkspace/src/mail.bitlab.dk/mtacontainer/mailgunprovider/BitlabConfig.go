package mailgunprovider
import "strconv"


func BitLabConfig(passphrase string) map[string]string {

	var mailGunConfig = make(map[string]string);
	mailGunConfig[MG_CNF_PASSPHRASE] = passphrase;
	// Encrypted API key, key gotten from MG-account using log-in at
	// https://mailgun.com/sessions/new. Now we can safely commit and push this
	// to GitHub
	mailGunConfig[MG_CNF_ENCRYPTED_APIKEY] =
	"59FZ4m/4sEY3viC7dSZKFYZchn3fqdSDubS+ZhMhyIXJXetKAAAAAAAAAAAAAAAA";
	mailGunConfig[MG_CNF_ENCRYPTED_APIKEY_LEN] = strconv.Itoa(36);
	mailGunConfig[MG_CNF_DOMAIN_TO_SERVE] = "mail.bitlab.dk";
	mailGunConfig[MG_CONF_HEALTH_NOTIFY_EMAIL] = "r@wz.gl";
	mailGunConfig[MG_CNF_ROUTE_ACTION_ON_INCOMING_MAIL]="forward(\"https://mail.bitlab.dk:31415/msg\")";

	return mailGunConfig;

}