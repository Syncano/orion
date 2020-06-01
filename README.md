# Orion

Syncano Platform written in Go. Meant to completely replace it, currently used as a partial v3 API support.

[![CircleCI](https://circleci.com/gh/Syncano/orion.svg?style=svg)](https://circleci.com/gh/Syncano/orion)

## Dependencies

- Golang version 1.14.
- docker 17.03+ and docker-compose (`pip install docker-compose`).

## Testing

- Run `make test` to run code checks and all tests with coverage. This will require Go installed on host.
- During development it is very useful to run dashboard for tests through `goconvey`. Install and run through `make goconvey`.
- To run tests in container run: `make test-in-docker`.

## Starting locally

- Build executable binary by `make build` or `make build-in-docker`. They both do the same but the first one requires dependencies to installed on local machine. Second command will automatically fetch all dependencies in docker container.
- Rebuild the image by `make docker`.
- Run `make start` to spin up 1 load balancer and 1 worker instance.

## Deployment

- You need to first build a static version and a docker image. See first two steps of **Starting locally** section.
- Make sure you have a working `kubectl` installed and configured. During deployment you may also require `gpg` (gnupg) and `jinja2-cli` (`pip install jinja2-cli[yaml]`).
- Run `make deploy-staging` to deploy on staging or `make deploy-production` to deploy on production.
