# Build Aggregator

A Bitrise enabled app to trigger Bitrise builds and collect their results.

## Usage

After cloning the repository, you have to create a `.bitrise.secrets.yml` and specify the following variables:

  - `BUILD_AGGREGATOR_SLACK_WEBHOOK`: Slack webhook URL accepting messages
  - `BITRISE_API_TOKEN`: Bitrise personal access token
  - `GITHUB_ACCESS_TOKEN`: GitHub personal access token
  - `CHANNEL`: Slack channel to send build outcome messages
  - `STEPLIB_SPEC_URL`: Steplib URL (eg: `http://localhost:8088`)
  - `GITHUB_ORGS`: comma separated list of GitHub organisations to filter the steplib steps (eg: `octocat,lszucs`)

That's it! Now you can configure whatever workflow suits you and use `bitrise run` to run it.

Two run arguments are required:

  -  `--steplib-spec-url`: Steplib URL (eg: `http://localhost:8088`)
  - `--github-orgs`: comma separated list of GitHub organisations to filter the steplib steps (eg: `octocat,lszucs`)
  - `--batch-size`: integer value, specifies how many builds will be processed in one batch

You can set these arguments in `.bitrise.secrets.yml` and have them automatically passed when using the `bitrise` cli to run the workflow.

## Development and testing

Inside the `test` folder you will find a `spec.json` and a `steplib.go` file. The former is a dummy Bitrise Step Library specification, the latter is a webserver to expose the spec.

Just type `go run ./. localhost:8088` and you have a local steplib. You can set the URL for `STEPLIB_SPEC_URL` inside `.bitrise.secrets.yml`.

## Running without Bitrise

The app is fully runnable outside of Bitrise. Just compile and run. You will have to specify environment variables manually or pass the relevant flags when calling the binary though.
