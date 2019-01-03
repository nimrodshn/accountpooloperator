/*
Copyright (c) 2018 Red Hat, Inc.

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

package accountprovisioner

import (
	"fmt"
	"time"

	"k8s.io/client-go/rest"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/organizations"
	// nolint
	accountpool "gitlab.cee.redhat.com/service/uhc-clusters-service/pkg/accountpooloperator/pkg/apis/accountpooloperator/v1"
	// nolint
	clientset "gitlab.cee.redhat.com/service/uhc-clusters-service/pkg/accountpooloperator/pkg/client/clientset/versioned"
)

var (
	pollAWSPeriod = time.Minute * 10
)

type AWSAccountProvisioner struct {
	Config *rest.Config
}

func (a *AWSAccountProvisioner) ProvisionAccount(
	account *accountpool.AWSAccount,
	creds map[string]string,
	stopCh <-chan struct{}) {
	// Handle any account errors that might occurre.
	errCh := make(chan error)
	go a.handleErrors(account, errCh, stopCh)

	err := a.validateCredentialsExist(creds)
	if err != nil {
		errCh <- err
		close(errCh)
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(creds["access_key_id"],
			creds["secret_access_key"], ""),
	})
	if err != nil {
		errCh <- err
		// close the error channel as the creation failed so there is no
		// need to keep monitoring errors.
		close(errCh)
		return
	}
	svc := organizations.New(sess)
	createAccountInput := &organizations.CreateAccountInput{
		AccountName: aws.String(account.Spec.AccountName),
		// TODO: Should be the SRE-P email?
		Email: aws.String(account.Spec.Email),
	}

	result, err := svc.CreateAccount(createAccountInput)
	if err != nil {
		a.AWSErr(err, errCh)
		// close the error channel as the creation failed so there is no
		// need to keep monitoring errors.
		close(errCh)
		return
	}

	statusID := result.CreateAccountStatus.Id

	// Run watchAccountState worker.
	go a.watchAndUpdateAccountStatus(statusID, account, creds, errCh, stopCh)
}

// watchAndUpdateAccountStatus periodically updates the account status field according to the status
// provided by AWS.
func (a *AWSAccountProvisioner) watchAndUpdateAccountStatus(
	statusID *string,
	account *accountpool.AWSAccount,
	creds map[string]string,
	errCh chan<- error,
	stopCh <-chan struct{}) {
	ticker := time.NewTicker(pollAWSPeriod)
	for {
		select {
		case <-ticker.C:
			a.updateAccountStatus(statusID, account, creds, errCh)
		case <-stopCh:
			return
		}
	}
}

// updateAccountStatus fetches the provisioned account state and updates the corresponding
// AWSAccount object to match the upstream state.
func (a *AWSAccountProvisioner) updateAccountStatus(
	statusID *string,
	account *accountpool.AWSAccount,
	creds map[string]string,
	errCh chan<- error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(creds["access_key_id"],
			creds["secret_access_key"], ""),
	})
	if err != nil {
		errCh <- err
	}
	svc := organizations.New(sess)

	// fetch account status
	describeAccountInput := &organizations.DescribeCreateAccountStatusInput{
		CreateAccountRequestId: statusID,
	}
	result, err := svc.DescribeCreateAccountStatus(describeAccountInput)
	if err != nil {
		a.AWSErr(err, errCh)
	}

	// align CRD status with to the corresponding AWS status
	var state accountpool.AccountStatus
	switch *result.CreateAccountStatus.State {
	case organizations.CreateAccountStateInProgress:
		state = accountpool.StatusInstalling
	case organizations.CreateAccountStateSucceeded:
		state = accountpool.StatusReady
	case organizations.CreateAccountStateFailed:
		state = accountpool.StatusError
	default:
		state = accountpool.StatusPending
	}

	// Update CRD status according to the upstream status
	if account.Spec.Status != state {
		account.Spec.Status = state
		clientset, err := clientset.NewForConfig(a.Config)
		if err != nil {
			errCh <- err
		}
		_, err = clientset.AccountpooloperatorV1().AWSAccounts(account.Namespace).Update(account)
		if err != nil {
			errCh <- err
		}
	}
}

func (a *AWSAccountProvisioner) AWSErr(err error, errCh chan<- error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case organizations.ErrCodeAccessDeniedException:
			errCh <- fmt.Errorf(organizations.ErrCodeAccessDeniedException, aerr.Error())
		case organizations.ErrCodeAWSOrganizationsNotInUseException:
			errCh <- fmt.Errorf(organizations.ErrCodeAWSOrganizationsNotInUseException, aerr.Error())
		case organizations.ErrCodeConcurrentModificationException:
			errCh <- fmt.Errorf(organizations.ErrCodeConcurrentModificationException, aerr.Error())
		case organizations.ErrCodeConstraintViolationException:
			errCh <- fmt.Errorf(organizations.ErrCodeConstraintViolationException, aerr.Error())
		case organizations.ErrCodeInvalidInputException:
			errCh <- fmt.Errorf(organizations.ErrCodeInvalidInputException, aerr.Error())
		case organizations.ErrCodeFinalizingOrganizationException:
			errCh <- fmt.Errorf(organizations.ErrCodeFinalizingOrganizationException, aerr.Error())
		case organizations.ErrCodeServiceException:
			errCh <- fmt.Errorf(organizations.ErrCodeServiceException, aerr.Error())
		case organizations.ErrCodeTooManyRequestsException:
			errCh <- fmt.Errorf(organizations.ErrCodeTooManyRequestsException, aerr.Error())
		case organizations.ErrCodeCreateAccountStatusNotFoundException:
			fmt.Println(organizations.ErrCodeCreateAccountStatusNotFoundException, aerr.Error())
		default:
			errCh <- fmt.Errorf(aerr.Error())
		}
	} else {
		// Some other error occurred - Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		errCh <- fmt.Errorf(err.Error())
	}
}

func (a *AWSAccountProvisioner) handleErrors(
	account *accountpool.AWSAccount,
	errCh <-chan error,
	stopCh <-chan struct{}) {
	for {
		select {
		case err, ok := <-errCh:
			// channel closed - shut down errorhandling thread.
			if !ok {
				return
			}
			// TODO: Possibly contact SRE-P team here to check the error
			fmt.Printf("Some error occured while trying to create account %v: %v", account.Name, err)
		case <-stopCh:
			return
		}
	}
}

func (a *AWSAccountProvisioner) validateCredentialsExist(creds map[string]string) error {
	if creds["secret_access_key"] == "" || creds["access_key_id"] == "" {
		return fmt.Errorf("an error occurred: user did not supply aws credentials for " +
			"the root master account of an organization.")
	}
	return nil
}