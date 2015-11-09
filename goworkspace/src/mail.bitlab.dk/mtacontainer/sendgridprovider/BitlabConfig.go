package sendgridprovider
import "strconv"


func BitLabConfig(passphrase string) map[string]string {

	var sendGridConfig = make(map[string]string);
	sendGridConfig[SG_CNF_PASSPHRASE] = passphrase;
	// Encrypted API key, key gotten from MG-account using log-in at
	// https://mailgun.com/sessions/new. Now we can safely commit and push this
	// to GitHub
	sendGridConfig[SG_CNF_ENC_API_KEY] =
	"3/MOqxej0DsAmgXZOCcHeMZo8n617MaG/avLPwRUwtL8WcpABfBvGwOUatYOtCW+wwK3q+c90Wp148C/IxyWXltrdVp4AAAAAAAAAAAAAAA=";
	sendGridConfig[SG_CNF_API_KEY_LEN] = strconv.Itoa(69);
	sendGridConfig[SG_CNF_API_USER] = "rasmuswl";

	return sendGridConfig;

}