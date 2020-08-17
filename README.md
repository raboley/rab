# rab
golang cli utility

## Installation

Using the github secrets command requires to install sodiumlib
`brew install libsodium`

after that you can go get it

```shell script
go get github.com/raboley/rab
```

## Usage

### Github

#### Add secrets

You can add secrets to a repo by using 

```shell script
export MY_SECRET="SECRET VALUE"
rab github add secret MY_SECRET --owner raboley --repo rab
```

Secrets should be pulled from environment variables.

you can also add multiple secrets

```shell script
export SECRET1="value 1"
export SECRET2="value 2"
rab github add secrets SECRET1,SECRET2 --owner raboley --repo rab
```

I generally need a lot of terraform and azure secrets in my actions so I might do something like this:

```shell script
export REPO="aks-infra"
export OWNER="raboley"
rab github add secrets API_GITHUB_TOKEN,ARM_CLIENT_ID,ARM_CLIENT_SECRET,ARM_SUBSCRIPTION_ID,ARM_TENANT_ID,TF_API_TOKEN,TF_GITHUB_TOKEN --repo $REPO --owner $OWNER
```