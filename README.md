# Speakerbob

---

Speakerbob is s a distributed sound board application.  Anyone can upload a short audio clip and play it!  The catch is the sound will play on any device with the speakerbob webpage open!  Annoy your co-workers, give yourself a theme song when you come into office, or just add sound affects to your next online gaming session with friends.

## Installation

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

## API

Want to automate sount effects for your life? Checkout the [api docs](https://github.com/paynejacob/speakerbob/tree/master/docs) to get started.

## Contributing

Have an idea for a feature? Found a bug? Please [create an issue.](https://github.com/paynejacob/speakerbob/issues/new)

Pull requests are always welcome!  If you want to resolve an issue, please make sure it is not assigned to anyone before starting on it.  The assignee's pr will always be given favor.

## License

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

