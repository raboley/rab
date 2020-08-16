/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	sodium "github.com/GoKillers/libsodium-go/cryptobox"
	"github.com/google/go-github/v32/github"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
	"os"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"new", "update"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		add(cmd, args)
	},
}

func init() {
	githubCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	addCmd.Flags().String("repo", "", "The name of the repo you want to add the secret to ex. rab")
	addCmd.Flags().String("owner", "", "The owner/user of the repo you want to add the secret to, ex. raboley")
}

func add(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		_ = cmd.Help()
		os.Exit(0)
	}
	var err error
	var addedThing string

	switch args[0] {
	case "secret":
		addedThing, err = addSecret(args, cmd.Flags())
	default:
		_ = cmd.Help()
		os.Exit(0)
	}

	if err != nil {
		fmt.Println(err)
		_ = cmd.Help()
		os.Exit(0)
	}

	fmt.Println("Added/Updated ", args[0], addedThing)
}

func addSecret(args []string, flags *pflag.FlagSet) (string, error) {
	if len(args) < 2 {
		err := errors.New("not enough args for command error")
		return "", err
	}

	// Get secrets from environment variables because that is better than passed
	// over command line
	secretName := args[1]
	secretValue := os.Getenv(secretName)
	if secretValue == "" {
		return "", errors.New(fmt.Sprintf(`secret with name: %s not defined as an environment variable, 
please export the variable as an enviornment variable so it can be read in

ex:
	export %s="secret value"
`, secretName, secretName))
	}

	owner, err := flags.GetString("owner")
	if err != nil {
		return "", err
	}
	if owner == "" {
		return "", errors.New("required flag --owner was not passed")
	}

	repo, err := flags.GetString("repo")
	if err != nil {
		return "", err
	}
	if repo == "" {
		return "", errors.New("required flag --repo was not passed")
	}

	createdSecret, err := AddRepoSecret(owner, repo, secretName, secretValue)
	if err != nil {
		return "", err
	}

	return createdSecret, nil
}

// GithubAuth reads an api token from the environment
// expecting API_GITHUB_TOKEN and authenticates with github api
// and returns an authenticated client.
func GithubAuth() (context.Context, *github.Client) {
	token := os.Getenv("API_GITHUB_TOKEN")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	return ctx, client
}

// AddRepoSecret will add a secret value to a given github repo for a given owner
// It encrypts the secret using sodium before sending the secret to github api
// therefore requires libsodium to be installed on the machine running this code
// https://formulae.brew.sh/formula/libsodium
// Github is very picky over formats things need to be to be sent, and not very descriptive in how
// they will be sent to the user
// To get a secret uploaded you need to get the public key of the repo that will be receiving the
// secret. This key is used to encrypt the secret before transport on the sending side, and then decrypted by github.
// Once you have the public key you need to encrypt the secret with sodiumlib before sending it.
// The public key comes base64 encoded, and sodiumlib expects it to not be base64 encoded so you need
// to decode the public key before using it.
// Once the public key is decoded you need to convert the string secret into bytes.
// once you have the public key decoded, and the secret string in bytes you can encrypt it using
// sodium.CryptoBoxSeal
// That will produce the correctly encrypted secret as bytes, but you need to then convert it to
// a base64 encoded string. After doing that you can use that base64 encoded string as the encrypted value
// to be part of the github.EncodedSecret type.
// The name is the string (no encoding, or base64 needed) of the secret name that will appear in github secrets
// then the KeyID will be the public key of the repo's ID, which is gettable from the public key's GetKeyID method.
// Finally you can pass that object in and have it be created or updated in github.
func AddRepoSecret(owner string, repo string, secretName string, secretValue string) (string, error) {
	ctx, client := GithubAuth()
	publicKey, _, err := client.Actions.GetRepoPublicKey(ctx, owner, repo)
	if err != nil {
		return "", err
	}

	encryptedSecret, err := encryptSecretWithPublicKey(publicKey, secretName, secretValue)
	if err != nil {
		return "", err
	}

	_, err = client.Actions.CreateOrUpdateRepoSecret(ctx, owner, repo, encryptedSecret)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Actions.CreateOrUpdateRepoSecret returned error: %v", err))
	}

	return secretName, nil
}

func encryptSecretWithPublicKey(publicKey *github.PublicKey, secretName string, secretValue string) (*github.EncryptedSecret, error) {
	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey.GetKey())
	if err != nil {
		return nil, errors.New(fmt.Sprintf("base64.StdEncoding.DecodeString was unable to decode public key: %v", err))
	}

	secretBytes := []byte(secretValue)
	encryptedBytes, exit := sodium.CryptoBoxSeal(secretBytes, decodedPublicKey)
	if exit != 0 {
		return nil, errors.New("sodium.CryptoBoxSeal exited with non zero exit code")
	}

	encryptedString := base64.StdEncoding.EncodeToString(encryptedBytes)
	keyID := publicKey.GetKeyID()
	encryptedSecret := &github.EncryptedSecret{
		Name:           secretName,
		KeyID:          keyID,
		EncryptedValue: encryptedString,
	}
	return encryptedSecret, nil
}
