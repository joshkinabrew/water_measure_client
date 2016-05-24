# Water Measure Client

This was meant as a project using a MaxBotix sensor to measure an underground cistern at an interval to ensure the water level was at an appropriate level. This client does an HTTP POST to a server to send the reading.

### Compiling for Raspberry Pi
It should be as simple as `GOOS=linux GOARCH=arm GOARM=7 go build`

### How to use
Once this is built and the binary copied to the Raspberry Pi, use like so:

`./water_measure_client "https://my-server.mydomain.com/telemetry/endpoint.json?auth=token"`
