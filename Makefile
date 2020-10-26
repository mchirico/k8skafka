
ifndef $(GOPATH)
  export GOPATH=${HOME}/gopath
  ${shell mkdir -p ${GOPATH}}
endif

ifndef $(GOBIN)
  export GOBIN=${GOPATH}/bin
endif


.PHONY: docker-build
docker-build:
	docker build --no-cache -t gcr.io/mchirico/gopub:test -f Dockerfile_pub .
	docker build --no-cache -t gcr.io/mchirico/readtemp:test -f Dockerfile_readtemp .


.PHONY: load
load:
	kind load docker-image gcr.io/mchirico/readtemp:test
	kind load docker-image gcr.io/mchirico/gopub:test


.PHONY: cert-manager
cert-manager:
	go get k8s.io/kubernetes || true
	cd ${GOPATH}/src/k8s.io/kubernetes && git checkout v1.19.2
	go get sigs.k8s.io/kind
	${GOPATH}/bin/kind build node-image --image=master
	${GOPATH}/bin/kind delete cluster
	${GOPATH}/bin/kind create cluster --config calico/kind-calico-pvc.yaml
	kubectl apply -f calico/ingress-nginx.yaml
	kubectl apply -f calico/tigera-operator.yaml
	kubectl apply -f calico/calicoNetwork.yaml
	kubectl apply -f calico/calicoctl.yaml
	kubectl apply -f calico/cert-manager.yaml
	export VERSION=0.17.1

#	kubectl kudo init
#	kubectl kudo install zookeeper
#	kubectl kudo install kafka






.PHONY: push
push:
	docker push gcr.io/mchirico/k8skafka:test


build:
	go build -v .

run:
	docker run --rm -it -p 3000:3000  gcr.io/mchirico/k8skafka:test
