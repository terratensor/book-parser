check:
	golangci-lint run -v

check-fix:
	golangci-lint run --fix

docker-build: build-kob-library-parser

build-kob-library-parser:
	DOCKER_BUILDKIT=1 docker --log-level=debug build --pull --build-arg BUILDKIT_INLINE_CACHE=1 \
		--target builder \
		--cache-from ${REGISTRY}/kob-library-parser:cache-builder \
		--tag ${REGISTRY}/kob-library-parser:cache-builder \
		--file ./docker/kob/Dockerfile .

	DOCKER_BUILDKIT=1 docker --log-level=debug build --pull --build-arg BUILDKIT_INLINE_CACHE=1 \
    	--cache-from ${REGISTRY}/kob-library-parser:cache-builder \
    	--cache-from ${REGISTRY}/kob-library-parser:cache \
    	--tag ${REGISTRY}/kob-library-parser:cache \
    	--tag ${REGISTRY}/kob-library-parser:${IMAGE_TAG} \
    	--file ./docker/kob/Dockerfile .

push-build-cache:
	docker push ${REGISTRY}/kob-library-parsercache-builder
	docker push ${REGISTRY}/kob-library-parser:cache

push:
	docker push ${REGISTRY}/kob-library-parser:${IMAGE_TAG}

kob-library-build:
	docker --log-level=debug build --pull --file=docker/kob/Dockerfile --tag=${REGISTRY}/kob-library-parser:${IMAGE_TAG} .
