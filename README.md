# imgopt

Image resizer API. Give it a URL and it'll upload optimized versions to S3. Built with Go and libvips. Still largely a work in progress.

## how to use

send a post request to `/r` with a json body containing the parameters in [models/request.go](./models/request.go)