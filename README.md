# Orion

[![CircleCI](https://circleci.com/gh/Syncano/orion.svg?style=svg&circle-token=439f828713b1869aa98396570b0aa9e07298fa31)](https://circleci.com/gh/Syncano/orion/tree/devel)

## Dependencies

- Golang version 1.10.
- docker 1.13+ and docker-compose (`pip install docker-compose`).
- Run `make devdeps` to install all compile, testing and development dependencies.

## Testing

- Run `make test` to run code checks and all tests with coverage. This will require Go installed on host.
- During development it is very useful to run dashboard for tests through `goconvey`. Install and run through `make goconvey`.
- Whole project sources are meant to be put in $GOPATH/src path. This is especially important during development.
- To run tests in container run: `make test-in-docker`.
- There are two lints available. `make lint` runs only fast linters on code, which should report most of potential issues. During CI `make flint` is run which is much more resource-intensive and time-consuming but is also more in depth.

## Starting locally

- Build a static version of executable binary by `make build-static` or `make build-in-docker`. They both do the same but the first one requires dependencies to installed on local machine. Second command will automatically fetch all dependencies in docker container.
- Rebuild the image by `make docker`.
- Run `make start` to spin up 1 load balancer and 1 worker instance.

## Deployment

- You need to first build a static version and a docker image. See first two steps of **Starting locally** section.
- Make sure you have a working `kubectl` installed and configured.
- Run `make deploy-staging` to deploy on staging or `make deploy-production` to deploy on production.
