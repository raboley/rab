package cmd

import (
	"github.com/google/go-github/v32/github"
	"testing"
)

func TestEncryptSecretWithPublicKey(t *testing.T) {
	secretName := "SOME_SECRET"
	secretValue := "this is secret"

	keyID := "568250167242549743"
	base64PublicKey := "zq9uEuErfVboCkx+VRGeG9VdoIxox1h5FmDtWbNsqGs="
	githubPublicKey := github.PublicKey{
		KeyID: &keyID,
		Key:   &base64PublicKey,
	}

	// an example of what the output would be, but can't verify the output since
	// the encrypted value will change each run.
	//expectedSecret := github.EncryptedSecret{
	//	Name:                  secretName,
	//	KeyID:                 keyID,
	//	EncryptedValue:        "vnDXV3V5Etys/1gfkH+d1x+50hZc4GzL1VridovMxXGbI9u/9IkSvA7ZR5EsGY7ynmM9S8eTFj8B1Lu6mH0=",
	//}

	_, err := encryptSecretWithPublicKey(&githubPublicKey, secretName, secretValue)
	if err != nil {
		t.Errorf(err.Error())
	}

	//assert.Equal(t, &expectedSecret.Name, encryptedSecret.Name)

}
