# messenger

We only support mongoDB and LINE SDK temporarily. Please register the webhook url`https://YOUR_DOMAIN/webhook` to your `LINE Developers Console`.

`Help` If you don't have own domain and you can use `ngrok` to generate a public url for your localhost.

## Get Started

### Configuration

You can use config file or CLI flags to set the configuration.

`Notice` the config file should be placed in the root path.

```yaml
line:
  secret: Channel Secret is retrieved from LINE Developers Console.
  token: Channel Access Token is retrieved from LINE Developers Console.
server:
  port: "server port (default: 8080)"
db:
  user: the user of mongodb
  password: the user password of mongodb
  host: the host of mongodb
  port: the port of mongodb
```

`make help` will list what flags support.

### Installation

We use `mockgen` to generate mocks for testing so please install `mockgen` and use `make generate` to generate them.

### Usage

If you don't have mongodb instance and you can use container instead. Please refer to the container configuration with docker-compose.yaml or use `docker-compose up` / `make run-db`.

Use `go run main.go` or `make run` to launch the messenger server.

## API

Please refer to `openapi.yaml`.
