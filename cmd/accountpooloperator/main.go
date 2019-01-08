// Copyright 2018 Nimrod Shneor <nimrodshn@gmail.com>
// and other contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"github.com/nimrodshn/accountpooloperator/pkg/accountprovisioner"
	"github.com/nimrodshn/accountpooloperator/pkg/apis/accountpooloperator/v1"
	"github.com/nimrodshn/accountpooloperator/pkg/controllers"
	"github.com/nimrodshn/accountpooloperator/pkg/signals"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"log"
)

var kubeconfig string

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to Kubernetes config file for the kubernetes cluster wheren accountpoolmanager will be deployed.")
	flag.Parse()
}

func main() {
	var config *rest.Config
	var err error

	stopCh := signals.SetupSignalHandler()

	if kubeconfig == "" {
		log.Printf("using in-cluster configuration")
		config, err = rest.InClusterConfig()
	} else {
		log.Printf("using configuration from '%s'", kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if err != nil {
		panic(err)
	}

	err = v1.AddToScheme(scheme.Scheme)
	if err != nil {
		panic(err)
	}

	AWSaccountProvisioner := &accountprovisioner.AWSAccountProvisioner{Config: config}
	accountControllerFactory, err := controllers.NewAWSAccountControllerFactory(
		config,
		AWSaccountProvisioner,
		stopCh)
	if err != nil {
		panic("An error occurred while trying to create the account informer factory.")
	}

	// Create the controller and run - this call is non-blocking
	accountControllerFactory.CreateControllerAndRun()

	accountPoolInformerFactory, err := controllers.NewAccountPoolInformerFactory(
		config,
		stopCh)
	if err != nil {
		panic("An error occurred while trying to create the account informer factory.")
	}

	// Create the controller and run - this call is non-blocking
	accountPoolInformerFactory.CreateInformerAndRun()

	<-stopCh
}
