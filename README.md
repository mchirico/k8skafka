![Go](https://github.com/mchirico/gokafka/workflows/Go/badge.svg)
[![codecov](https://codecov.io/gh/mchirico/goKafka/branch/main/graph/badge.svg?token=AU1KLS2WIJ)](https://codecov.io/gh/mchirico/gokafka)


# goKafka

The project code is tested with docker-compose; but, I thought it would be interesting to
run the final test in a Kubernetes environment.  Below are instructions for installing a
KinD cluster, with compiles Kubernetes 1.19 from source.

## Unit Tests:  docker-compose

Note, docker-compose is launched via Go's tests.  Reference Test Main https://github.com/mchirico/gokafka/blob/74815c14058abd88ae84f49ec8df554ae2fd74c6/pkg/utils/utils_test.go#L14

The docker-compose file for testing can be found [here](https://github.com/mchirico/gokafka/blob/74815c14058abd88ae84f49ec8df554ae2fd74c6/compose/docker-compose.yml#L22).  Please note the default port has been changed to 29099.

```bash
# Generate temperature readings
go test -race -v -coverprofile=coverage3.txt ./temperature/readings

# Raw Kafka pubSub functions
go test -race -v -coverprofile=coverage0.txt ./pkg/utils

# Raw functions combined or wrapped
go test -race -v -coverprofile=coverage1.txt ./pkg/wrapper

# Using wrapped functions
go test -race -v -coverprofile=coverage2.txt ./temperature/pubsub

# Testing temperature readings
go test -race -v -coverprofile=coverage3.txt ./temperature/readings

```




# Prerequisites for Full Integration Test (k8s environment)

## 1. Kind
[KinD](https://kind.sigs.k8s.io/) 
```
go get sigs.k8s.io/kind
```

## 2. Kubernetes Source
[k8s.io/kubernetes](https://github.com/kubernetes/kubernetes)
```
go get k8s.io/kubernetes
```

## 3. Kudo
[Kudo](https://kudo.dev/docs/cli/installation.html#cli-installation)
```
export VERSION=0.17.1
OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
wget -O kubectl-kudo 
wget -O kubectl-kudo https://github.com/kudobuilder/kudo/releases/download/v${VERSION}/kubectl-kudo_${VERSION}_${OS}_${ARCH}
chmod +x kubectl-kudo
# add to your path
sudo mv kubectl-kudo /usr/local/bin/kubectl-kudo

```

## 4. kubectl

```
curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl
chmod +x ./kubectl
sudo mv ./kubectl /usr/bin/kubectl
```

## 5. Docker
It is assumed you have docker installed and running.


## Go
Install Go.  Set GOPATH and GOBIN to your preferred environment.

# Installation

Assuming docker, Go, kubectl,  and Kudo are installed, clone the directory.


### Clone the repo
```bash
git clone https://github.com/mchirico/gokafka.git
cd gokafka 
```

### Build Kubernetes v1.19 from source

The following make command will download and install Kubernetes from source, into a KinD
cluster with cert-manager.  This cluster can utilize a PVC from `/pod-10g/`, although, it's
not currently needed for this demo.  

Build time is about 4 minutes.

```bash
make cert-manager
```
It takes another 3 to 4 minutes for the pods to be in a running state.


![image](https://user-images.githubusercontent.com/755710/97125844-8d70ef80-170b-11eb-9bf8-04f549503277.png)

### kudo init
```bash
kubectl kudo init
```

![image](https://user-images.githubusercontent.com/755710/97125970-f22c4a00-170b-11eb-9f88-f6d384caf787.png)


### Zookeeper

Install Zookeeper
```bash
kubectl kudo install zookeeper
```
![image](https://user-images.githubusercontent.com/755710/97126153-839bbc00-170c-11eb-8188-055307775634.png)


Wait until all the pods are in a running state.
![image](https://user-images.githubusercontent.com/755710/97126185-8e565100-170c-11eb-8008-997f141f8fe4.png)


### Kafka
```bash
kubectl kudo install kafka
```

Wait until all the pods are in a running state.
![image](https://user-images.githubusercontent.com/755710/97126369-015fc780-170d-11eb-9dd8-a4c07f76d4a0.png)


### Build Docker Images

Just run `make` to build *gcr.io/mchirico/gopub:test* and *gcr.io/mchirico/readtemp:test*

```bash
make
```

### Load Images into KinD

Private or local docker images need to be loaded into the KinD notes.  This can be done with *make load*.

```bash
make load
```

### Deployments

Deploy the publisher and the consumer pods...

```bash
cd k8s
kubectl apply -f pub-deployment.yaml
kubectl apply -f sub-deployment.yaml
```

It should be possible to follow the logs, once the pods are running.

![image](https://user-images.githubusercontent.com/755710/97127144-3bca6400-170f-11eb-9276-d09247c27a10.png)


```bash
k logs --follow node-sub-795fd88bfd-k4jk8
```

Below, example output with debugging information...
![image](https://user-images.githubusercontent.com/755710/97127219-64525e00-170f-11eb-9135-1923c5f89088.png)






















