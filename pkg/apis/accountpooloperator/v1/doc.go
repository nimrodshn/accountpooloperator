// +k8s:deepcopy-gen=package,register

// Package v1 is the v1 version of the API.
// +groupName=accountpooloperator.openshift.io

//nolint
//go:generate ../../../../../../vendor/k8s.io/code-generator/generate-groups.sh all gitlab.cee.redhat.com/service/uhc-clusters-service/pkg/accountpooloperator/pkg/client gitlab.cee.redhat.com/service/uhc-clusters-service/pkg/accountpooloperator/pkg/apis "accountpooloperator:v1"

package v1
