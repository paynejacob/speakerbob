# Speakerbob

---

Speakerbob is s a distributed sound board application.  Anyone can upload a short audio clip and play it!  The catch is the sound will play on any device with the speakerbob webpage open!  Annoy your co-workers, give yourself a theme song when you come into office, or just add sound affects to your next online gaming session with friends.  Anyone can upload anything, if a sound is played while one is already playing it will be queued to play after the current one completes. 

## Installation

Before you get started, you will need a s3 compatible api with tag support.

The simplest way to get started is with the docker container.

```shell
docker run -p 80:80 paynejacob/speakerbob:latest server
```

For more stable deployments use the helm chart.

```shell
$ git clone https://github.com/paynejacob/speakerbob.git
$ helm install --create-namespace -n speakerbob speakerbob speakerbob/charts/speakerbob
```

If you do not want to use docker or kubernetes you can download the binary for your OS from the [releases page](https://github.com/paynejacob/speakerbob/releases).

## API Usage

### Create a Sound
Sounds are created in two steps:

1. `POST /sound/` accepts a multipart form with a single field containing an audio file.  Make sure to include the file extension in the field name.
2. `PATCH /sound/<id>/` Using the id returned from the last step this endpoint accepts a json payload with the `name` and `nsfw` field.

Example:

```http request
POST /sound/
Content-Type: multipart/form-data

--2a8ae6ad-f4ad-4d9a-a92c-6d217011fe0f
Content-Disposition: form-data; name="sound.mp3"; filename="sound.mp3"
Content-Type: audio/mp3

<binary data>;
```
```http request
PATCH /sound/<id>/
Content-Type: application/json

{"name":"fart noise","nsfw":false}
```

### Play a Sound
To queue a sound for playback:

`PUT /play/sound/<id>/`

### List Sounds
To list all sounds:

`GET /sound/`

### Search Sounds
To search for sounds:

`GET /sound/?q=<query>`

## Contributing

Have an idea for a feature? Found a bug? Please [create an issue.](https://github.com/paynejacob/speakerbob/issues/new)

Pull requests are always welcome!  If you want to resolve an issue, please make sure it is not assigned to anyone before starting on it.  The assignee's pr will always be given favor.

## License

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

