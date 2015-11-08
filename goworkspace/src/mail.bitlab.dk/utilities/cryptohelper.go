// ------------------------------------------------------
//
// CRYPTO UTILS
//  [] In particular API keys for Amazon etc. are present as
//  cipher text constants in the code. Upon start up (in particular
//  for the mtaserver) these are decrypted by a password provided
//  on command-line. This is solely to ensure the API keys doesn't
//  end up on GitHub for public (mis)usage.
//
//
package utilities
import (
	"hash"
	"crypto/sha256"
	"encoding/hex"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

func HashStringToHex(str string) string {
	var randomOracle hash.Hash = sha256.New();
	randomOracle.Write([]byte(str));
	var bytes = randomOracle.Sum(nil);
	return hex.EncodeToString(bytes);
}

func ComputeAesKey(passphrase string) []byte {
	var randomOracle hash.Hash = sha256.New();
	randomOracle.Write([]byte(passphrase));
	// TODO(rwz): Remove this debug println
	println("[keygen] passphrase as bytes:\n" + hex.Dump([]byte(passphrase)));
	var hashedPassword = randomOracle.Sum(nil);
	var rawKey = hashedPassword[:16];

	return rawKey;
}

func EncryptApiKey(apiKey string, passphrase string) string {

	var plaintext = []byte(apiKey);
	var aesKey = ComputeAesKey(passphrase);

	var aesBlockCipher, aesBlockCipherErr = aes.NewCipher(aesKey);
	if (aesBlockCipherErr != nil) {
		return "";
	}

	var iv = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
	var aesBlockCipherCounterMode = cipher.NewCTR(aesBlockCipher, iv);

	var cipherText = make([]byte, ((len(plaintext) / 16 + 1) * 16));
	aesBlockCipherCounterMode.XORKeyStream(cipherText, plaintext);

	return base64.StdEncoding.EncodeToString(cipherText);
}

//
// From MailGun DashBoard we see that the secret api key is 36 Ascii-characters
//
func DecryptApiKey(passphrase string, encryptedKeyB64 string, apiKeyLen int) string {

	var cipherText, cipherTextErr = base64.StdEncoding.DecodeString(encryptedKeyB64);
	if cipherTextErr != nil {
		return "";
	}

	var aesKey = ComputeAesKey(passphrase);
	var aesBlockCipher, aesBlockCipherErr = aes.NewCipher(aesKey);
	if (aesBlockCipherErr != nil) {
		return "";
	}

	var iv = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
	var aesBlockCipherCounterMode = cipher.NewCTR(aesBlockCipher, iv);

	var plaintext = make([]byte, len(cipherText));
	aesBlockCipherCounterMode.XORKeyStream(plaintext, cipherText);


	return string(plaintext[:apiKeyLen]);
}
