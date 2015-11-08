package sendgridprovider
import "strconv"


func BitLabConfig(passphrase string) map[string]string {

	var sendGridConfig = make(map[string]string);
	sendGridConfig[SG_CNF_PASSPHRASE] = passphrase;
	// Encrypted API key, key gotten from MG-account using log-in at
	// https://mailgun.com/sessions/new. Now we can safely commit and push this
	// to GitHub
	sendGridConfig[SG_CNF_ENC_API_KEY] =
	"59FZ4m/4sEY3viC7dSZKFYZchn3fqdSDubS+3/MOpi6a8S8Ljg3ZKFYHSINQ6kyp2P7I7avoOQFUoNLNH9d6VKlNZjW6H4QLvBW9wADNqdcj8QtCwsCNUlK4JktbT3sdAAAAAAAAAAAAAAA=";
	sendGridConfig[SG_CNF_API_KEY_LEN] = strconv.Itoa(69);
	sendGridConfig[SG_CNF_API_USER] = "rasmuswl";

	return sendGridConfig;

}