DOCKERNAME = samfalkner/cc-to-stripe
EXECUTABLE = cc-to-stripe

build $(EXECUTABLE): clean
	go build

docker: export GOOS=linux
docker: export GOARCH=amd64
docker: clean
	packr build
	docker build -t samfalkner/cc-to-stripe .

push: docker
	docker push samfalkner/cc-to-stripe

clean: FORCE
	rm -f $(EXECUTABLE)

FORCE:
