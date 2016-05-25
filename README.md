# Water Measure Client

This was created from a project I worked on using a MaxBotix sensor to measure an underground cistern at a specified interval to ensure the water level was at an appropriate level. This client does an HTTP POST to a server to send the reading.

## Features

* If the reading is not successfully sent to the server, it saves in a backup JSON file. The next time it is able to send a reading to the server it retrieves the backup file and sends each of the readings in the JSON file. This is helpful for when an internet connection on the Raspberry Pi goes down temporarily and you still want to save the readings.
* The client knows the correct format of a reading value and makes sure to not send an incorrect value to the server. It will retry up to 5 times to get a correct reading before failing to get the reading.

## Compiling for use on Raspberry Pi
It is as simple as running `GOOS=linux GOARCH=arm GOARM=7 go build`

## How to use
Once this project is built as per instructions above and the binary file is copied to the Raspberry Pi, use like so:

`./water_measure_client https://my-server.mydomain.com/telemetry/endpoint.json?auth=token`

Additionally you may add it in a Cron job to run at a specified interval like so:

`*/15 * * * * pi /home/pi/water_measure https://my-server.mydomain.com/telemetry/endpoint.json?auth=token >> /home/pi/water_measure.log 2>&1`

(This cron job example runs every 15 minutes but feel free to customize it to your liking)
