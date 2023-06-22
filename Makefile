check:
	golangci-lint run -v

check-fix:
	golangci-lint run --fix

docker-build: build-book-parser-common build-postgres

build-book-parser-common:
	DOCKER_BUILDKIT=1 docker --log-level=debug build --pull --build-arg BUILDKIT_INLINE_CACHE=1 \
		--target builder \
		--cache-from ${REGISTRY}/book-parser-common:cache-builder \
		--tag ${REGISTRY}/book-parser-common:cache-builder \
		--file ./docker/Dockerfile .

	DOCKER_BUILDKIT=1 docker --log-level=debug build --pull --build-arg BUILDKIT_INLINE_CACHE=1 \
    	--cache-from ${REGISTRY}/book-parser-common:cache-builder \
    	--cache-from ${REGISTRY}/book-parser-common:cache \
    	--tag ${REGISTRY}/book-parser-common:cache \
    	--tag ${REGISTRY}/book-parser-common:${IMAGE_TAG} \
    	--file ./docker/Dockerfile .

build-postgres:
	DOCKER_BUILDKIT=1 docker --log-level=debug build --pull --build-arg BUILDKIT_INLINE_CACHE=1 \
        --cache-from ${REGISTRY}/postgres-book-parser:cache \
        --tag ${REGISTRY}/book-parser-common:cache \
        --tag ${REGISTRY}/book-parser-common:${IMAGE_TAG} \
        --file docker/postgres/Dockerfile docker/postgres

push-build-cache:
	docker push ${REGISTRY}/book-parser-common:cache-builder
	docker push ${REGISTRY}/book-parser-common:cache
	docker push ${REGISTRY}/postgres-book-parser:cache

push:
	docker push ${REGISTRY}/book-parser-common:${IMAGE_TAG}
	docker push ${REGISTRY}/postgres-book-parser:${IMAGE_TAG}