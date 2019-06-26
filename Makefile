# Copyright 2018 The Ceph-CSI Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: all liveness

CONTAINER_CMD?=docker

LIVENESS_IMAGE_NAME=$(if $(ENV_LIVENESS_IMAGE_NAME),$(ENV_LIVENESS_IMAGE_NAME),quay.io/daniel_pivonka/liveness)
LIVENESS_IMAGE_VERSION=$(if $(ENV_LIVENESS_IMAGE_VERSION),$(ENV_LIVENESS_IMAGE_VERSION),canary)


all: push-image-liveness

liveness:
	if [ ! -d ./vendor ]; then dep ensure -vendor-only; fi
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o  _output/liveness ./cmd/

image-liveness: liveness
	cp _output/liveness .
	$(CONTAINER_CMD) build -t $(LIVENESS_IMAGE_NAME):$(LIVENESS_IMAGE_VERSION) .

push-image-liveness: image-liveness
		$(CONTAINER_CMD) push $(LIVENESS_IMAGE_NAME):$(LIVENESS_IMAGE_VERSION)

clean:
	go clean -r -x
	rm -f liveness
	rm -rf _output
