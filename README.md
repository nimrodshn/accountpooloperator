# AccountPoolOperator

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
    secret_access_key: "1234"
    access_key_id: "4567"
  pool_size: 2
```

**Please Note:** AccountPoolOperator currently only supports AWS cloud platform and is undergoing constant changes.

Using this file to create an account pool is as simple as: `kubectl create -f example_pool.yml`

To watch you're account pool one can use the following command: `kubectl get awsaccounts -w`

**A note about credentials:** The credentials passed in the body of `example_pool.yml` must be credentials
of an account with permissions to create accounts in an AWS organization (or the root account in that organization).

One can easily edit the pool size by editing the `accountpool` CRD using `kubectl edit accountpool example-pool`.

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



