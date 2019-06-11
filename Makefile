test-env-up:
	docker run -d --rm --name pg -p 5432:5432 -v $(shell pwd)/sql:/docker-entrypoint-initdb.d postgres:12-beta1-alpine
	sleep 10 # wait for pg to start up

test-env-down:
	docker rm -f pg

test-only:
	go test -v -count 1 -race -cover ./...

bench:
	go test -v -run Bench -bench=. ./...

test: test-env-up test-only bench test-env-down
