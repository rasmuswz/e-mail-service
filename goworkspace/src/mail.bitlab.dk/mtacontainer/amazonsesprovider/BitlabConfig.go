package amazonsesprovider
import "strconv"



func BitLabConfig(passphrase string) map[string]string {

	var awsSesConfig = make(map[string]string);
	awsSesConfig[AWS_CNF_PASSPHRASE] = passphrase;
	// Encrypted API key, key gotten from MG-account using log-in at
	// https://mailgun.com/sessions/new. Now we can safely commit and push this
	// to GitHub
	awsSesConfig[AWS_CNF_API_KEY_ID] = "AKIAJOWWAJXJ2PUSVCWA";
	awsSesConfig[AWS_CNF_ENC_SECRET_KEY] = "9cYT/Wij2EQ9kArEODo2SPsL+nyd9MX94OzrIR5hloSeVM17Xa9HOQAAAAAAAAAA";
	awsSesConfig[AWS_CNF_SECRET_KEY_LEN] = strconv.Itoa(40);

	return awsSesConfig;

}