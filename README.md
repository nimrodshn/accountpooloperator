# AccountPoolOperator
[![Build Status](https://travis-ci.com/nimrodshn/accountpooloperator.svg?token=8bTEfKi17WtMetLXgARz&branch=master)](https://travis-ci.com/nimrodshn/accountpooloperator.svg?token=8bTEfKi17WtMetLXgARz&branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/nimrodshn/accountpooloperator)](https://goreportcard.com/report/github.com/nimrodshn/accountpooloperator)

AccountPoolOperator offers a declerative API for users to create, maintain and monitor account pools across cloud platforms.

The declerative API is built on top of kuberenetes using kubernetes tooling.

## How to create a new account pool?

Simply create a new yaml file like the following:

```
> cat example_pool.yml
apiVersion: accountpooloperator.openshift.io/v1
kind: AccountPool
metadata:
  name: "example-pool"
spec:
  provider: "AWS"
  credentials:
    name: "aws-credentials"
  pool_size: 2
```


**Please Note:** AccountPoolOperator currently only supports AWS cloud platform and is undergoing constant changes.

Using this file to create an account pool is as simple as: `kubectl create -f example_pool.yml`

One can easily watch, get, edit etc. their accountpool using the `kubectl` tooling.

## Credentials
The `spec.credentials.name` field is a reference to a secret which holds the **base64 encoded** AWS credentials: These must be credentials
of an account with permissions to create accounts in an AWS organization (or the root account in that organization).

## Managing accounts
To mark an account as "unavailable" simply change the "available" label on the awsaccount CRD:
```
kubectl label --overwrite awsaccount someaccount "available = false"
```
Or delete the awsaccount CRD altogether.

**Please note:** deleting the CRD does **NOT** delete the provisioned AWS accounts along with any artifacts associated with them - that can only be done manually using the AWS console.

## Deploying accountpooloperator
Accountpooloperator can built and deployed using the following methods:

1. The recommended way is to run Accountpooloperator as a pod inside a kubernetes cluster using the deployment file, by running - `kubectl create -f deployment.yml`.
2. Accountpooloperator can also be built and run as a binary using the `make` command.
3. A third way is to run it as docker container by running `make image` followed by `docker run nimrodshn/accountpoolopreator:latest --kubeconfig=...`.



