IMAGE=joshuapok/joshuahomelab
TAG=text_editor

build:
	docker build -t ${IMAGE}:${TAG} .

push:
	docker push ${IMAGE}:${TAG}
