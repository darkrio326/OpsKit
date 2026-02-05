.PHONY: docker-kylin-e2e

ROOT := $(abspath ../..)
OUTPUT ?= $(ROOT)/.tmp/generic-e2e-kylin-v10
IMAGE ?= kylinv10/kylin:b09
DOCKER_PLATFORM ?=
DRY_RUN ?= 0

docker-kylin-e2e:
	@OUTPUT="$(OUTPUT)" IMAGE="$(IMAGE)" DOCKER_PLATFORM="$(DOCKER_PLATFORM)" DRY_RUN="$(DRY_RUN)" ./run-af-kylin-v10-docker.sh
