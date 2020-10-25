
docker-build:
	docker build --no-cache -t gcr.io/mchirico/gopub:test -f Dockerfile_pub .
	docker build --no-cache -t gcr.io/mchirico/readtemp:test -f Dockerfile_readtemp .

load:
	kind load docker-image gcr.io/mchirico/readtemp:test
	kind load docker-image gcr.io/mchirico/gopub:test


push:
	docker push gcr.io/mchirico/goKafka:test

build:
	go build -v .

run:
	docker run --rm -it -p 3000:3000  gcr.io/mchirico/goKafka:test
