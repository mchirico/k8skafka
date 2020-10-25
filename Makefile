
docker-build:
	docker build --no-cache -t gcr.io/mchirico/goKafka:test -f Dockerfile .

push:
	docker push gcr.io/mchirico/goKafka:test

build:
	go build -v .

run:
	docker run --rm -it -p 3000:3000  gcr.io/mchirico/goKafka:test