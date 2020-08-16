package main

import (
	"github.com/raboley/rab/cmd"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetSecret(t *testing.T) {
	repo := "rab"
	owner := "raboley"

	ctx, client := cmd.GithubAuth()

	name := "TEST_READ"
	secret, response, err := client.Actions.GetRepoSecret(ctx, owner, repo, name)
	if err != nil {
		t.Log(response)
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, name, secret.Name)
}

func TestAddSecret(t *testing.T) {
	repo := "rab"
	owner := "raboley"
	secretName := "TEST_ADD"
	secretValue := "some secretValue"

	_, err := cmd.AddRepoSecret(owner, repo, secretName, secretValue)
	if err != nil {
		t.Errorf(err.Error())
	}

	ctx, client := cmd.GithubAuth()
	secret, response, err := client.Actions.GetRepoSecret(ctx, owner, repo, secretName)
	if err != nil {
		t.Log(response)
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, secretName, secret.Name)

	response, err = client.Actions.DeleteRepoSecret(ctx, owner, repo, secretName)
	if err != nil {
		t.Log(response)
		t.Log(err)
		t.Fail()
	}
}
