#!/bin/bash

# tests which use KIND should have code which can use an existing cluster if there is one
GO111MODULE="on" /usr/local/go/bin/go install sigs.k8s.io/kind@v0.11.1
$GOPATH/bin/kind create cluster --name $KIND_TEST_CLUSTER_NAME --image artifactory.appsflyer.com:5000/kindest/node:v1.21.1 --config test/kind/config.yml
# the api server listens on the docker dind ip, so we need to direct it there
# the reason the name is kubernetes and not docker is because docker is not valid for ssl certificates, more details below
sed -i 's/0.0.0.0/kubernetes/' ~/.kube/config
# add kubernetes to the hosts file with the docker ip which is the kubernetes api server addrress
echo "$(grep docker /etc/hosts | awk '{print $1}') kubernetes" >> /etc/hosts

#this is the error message received when trying to use docker
#Post "https://docker:33141/api/v1/namespaces": x509: certificate is valid for af-deployment-test-cluster-control-plane, kubernetes, kubernetes.default,
#kubernetes.default.svc, kubernetes.default.svc.cluster.local, localhost, not docker