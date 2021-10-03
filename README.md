## public-ip

The server returns the public IP of the request.

Supported http and https. Requests are not logged.

### Example

Request:
`curl https://path-to-service.com`

Response:
`170.44.249.129`

### Deploy:

1. `go build .`

2. Run HTTP and HTTPS: `./public-ip /path/to/certificate /path/to/private/key`

or

2. Run only HTTP: `./public-ip`
