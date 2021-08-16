# drone-runner-kube

The `kubernetes` runner executes pipelines inside Kubernetes pods. This runner is an alternative to the docker runner and is optimize for teams running Drone on Kubernetes. This requires Drone server `1.6.0` or higher.

<!--
Documentation:<br/>
https://kubernetes-runner.docs.drone.io
-->

Technical Support:<br/>
https://discourse.drone.io

Issue Tracker and Roadmap:<br/>
https://trello.com/b/ttae5E5o/drone

## Release procedure

Run the changelog generator.

```BASH
docker run -it --rm -v "$(pwd)":/usr/local/src/your-app githubchangeloggenerator/github-changelog-generator -u drone-runners -p drone-runner-kube -t <secret github token>
```

You can generate a token by logging into your GitHub account and going to Settings -> Personal access tokens.

Next we tag the PR's with the fixes or enhancements labels. If the PR does not fufil the requirements, do not add a label.

Run the changelog generator again with the future version according to semver.

```BASH
docker run -it --rm -v "$(pwd)":/usr/local/src/your-app githubchangeloggenerator/github-changelog-generator -u drone-runners -p drone-runner-kube -t <secret token> --future-release v1.0.0
```

Create your pull request for the release. Get it merged then tag the release.
