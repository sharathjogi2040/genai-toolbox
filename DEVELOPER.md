# DEVELOPER.md

This document provides instructions for setting up your development environment
and contributing to the Toolbox project.

## Prerequisites

Before you begin, ensure you have the following:

1. **Databases:** Set up the necessary databases for your development
   environment.
1. **Go:** Install the latest version of [Go](https://go.dev/doc/install).
1. **Dependencies:** Download and manage project dependencies:

    ```bash
    go get
    go mod tidy
    ```

## Developing Toolbox

### Running from Local Source

1. **Configuration:** Create a `tools.yaml` file to configure your sources and
   tools. See the [Configuration section in the
   README](./README.md#Configuration) for details.
1. **CLI Flags:** List available command-line flags for the Toolbox server:

    ```bash
    go run . --help
    ```

1. **Running the Server:** Start the Toolbox server with optional flags. The
   server listens on port 5000 by default.

    ```bash
    go run .
    ```

1. **Testing the Endpoint:** Verify the server is running by sending a request
   to the endpoint:

    ```bash
    curl http://127.0.0.1:5000
    ```

## Testing

### Infrastructure

Toolbox uses both GitHub Actions and Cloud Build to run test workflows. Cloud
Build is used when Google credentials are required. Cloud Build uses test
project "toolbox-testing-438616".

### Linting

Run the lint check to ensure code quality:

```bash
golangci-lint run --fix
```

### Unit Tests

Execute unit tests locally:

```bash
go test -race -v ./...
```

### Integration Tests

#### Running Locally

1. **Environment Variables:** Set the required environment variables. Refer to
   the [Cloud Build testing configuration](./.ci/integration.cloudbuild.yaml)
   for a complete list of variables for each source.
    * `SERVICE_ACCOUNT_EMAIL`: Use your own GCP email.
    * `CLIENT_ID`: Use the Google Cloud SDK application Client ID. Contact
      Toolbox maintainers if you don't have it.
1. **Running Tests:** Run the integration test for your target source. Specify
   the required Go build tags at the top of each integration test file.

    ```shell
    go test -race -v ./tests/<YOUR_TEST_DIR>
    ```

    For example, to run the AlloyDB integration test:

    ```shell
    go test -race -v ./tests/alloydbpg
    ```

#### Running on Pull Requests

* **Internal Contributors:** Testing workflows should trigger automatically.
* **External Contributors:** Request Toolbox maintainers to trigger the testing
  workflows on your PR.

#### Test Resources

The following databases have been added as test resources. To add a new database
to test against, please contact the Toolbox maintainer team via an issue or PR.
Refer to the [Cloud Build testing
configuration](./.ci/integration.cloudbuild.yaml) for a complete list of
variables for each source.

* AlloyDB - setup in the test project
  * AI Natural Language ([setup
    instructions](https://cloud.google.com/alloydb/docs/ai/use-natural-language-generate-sql-queries))
    has been configured for `alloydb-a`-nl` tool tests
  * The Cloud Build service account is a user
* Bigtable - setup in the test project
  * The Cloud Build service account is a user
* BigQuery - setup in the test project
  * The Cloud Build service account is a user
* Cloud SQL Postgres - setup in the test project
  * The Cloud Build service account is a user
* Cloud SQL MySQL - setup in the test project
  * The Cloud Build service account is a user
* Cloud SQL SQL Server - setup in the test project
  * The Cloud Build service account is a user
* Couchbase - setup in the test project via the Marketplace
* DGraph - using the public dgraph interface <https://play.dgraph.io> for
  testing
* Memorystore Redis - setup in the test project using a Memorystore for Redis
  standalone instance
  * Memorystore Redis Cluster, Memorystore Valkey standalone, and Memorystore
    Valkey Cluster instances all require PSC connections, which requires extra
    security setup to connect from Cloud Build. Memorystore Redis standalone is
    the only one allowing PSA connection.
  * The Cloud Build service account is a user
* Memorystore Valkey - setup in the test project using a Memorystore for Redis
  standalone instance
  * The Cloud Build service account is a user
* MySQL - setup in the test project using a Cloud SQL instance
* Neo4j - setup in the test project on a GCE VM
* Postgres - setup in the test project using an AlloyDB instance
* Spanner - setup in the test project
  * The Cloud Build service account is a user
* SQL Server - setup in the test project using a Cloud SQL instance
* SQLite -  setup in the integration test, where we create a temporary database
  file

### Other GitHub Checks

* License header check (`.github/header-checker-lint.yml`) - Ensures files have
  the appropriate license
* CLA/google - Ensures the developer has signed the CLA:
  <https://cla.developers.google.com/>
* conventionalcommits.org - Ensures the commit messages are in the correct
  format. This repository uses tool [Release
  Please](https://github.com/googleapis/release-please) to create GitHub
  releases. It does so by parsing your git history, looking for [Conventional
  Commit messages](https://www.conventionalcommits.org/), and creating release
  PRs. Learn more by reading [How should I write my
  commits?](https://github.com/googleapis/release-please?tab=readme-ov-file#how-should-i-write-my-commits)

## Developing Documentation

### Running a Local Hugo Server

Follow these steps to preview documentation changes locally using a Hugo server:

1. **Install Hugo:** Ensure you have
   [Hugo](https://gohugo.io/installation/macos/) extended edition version
   0.146.0 or later installed.
1. **Navigate to the Hugo Directory:**

    ```bash
    cd .hugo
    ```

1. **Install Dependencies:**

    ```bash
    npm ci
    ```

1. **Start the Server:**

    ```bash
    hugo server
    ```

### Previewing Documentation on Pull Requests

#### Contributors

Request a repo owner to run the preview deployment workflow on your PR. A
preview link will be automatically added as a comment to your PR.

#### Maintainers

1. **Inspect Changes:** Review the proposed changes in the PR to ensure they are
   safe and do not contain malicious code. Pay close attention to changes in the
   `.github/workflows/` directory.
1. **Deploy Preview:** Apply the `docs: deploy-preview` label to the PR to
   deploy a documentation preview.

## Building Toolbox

### Building the Binary

1. **Build Command:** Compile the Toolbox binary:

    ```bash
    go build -o toolbox
    ```

1. **Running the Binary:** Execute the compiled binary with optional flags. The
   server listens on port 5000 by default:

    ```bash
    ./toolbox
    ```

1. **Testing the Endpoint:** Verify the server is running by sending a request
   to the endpoint:

    ```bash
    curl http://127.0.0.1:5000
    ```

### Building Container Images

1. **Build Command:** Build the Toolbox container image:

    ```bash
    docker build -t toolbox:dev .
    ```

1. **View Image:** List available Docker images to confirm the build:

    ```bash
    docker images
    ```

1. **Run Container:** Run the Toolbox container image using Docker:

    ```bash
    docker run -d toolbox:dev
    ```

## Developing Toolbox SDKs

Refer to the [SDK developer
guide](https://github.com/googleapis/mcp-toolbox-sdk-python/blob/main/DEVELOPER.md)
for instructions on developing Toolbox SDKs.

## Maintainer Information

### Team

Team, `@googleapis/senseai-eco`, has been set as
[CODEOWNERS](.github/CODEOWNERS). The GitHub TeamSync tool is used to create
this team from MDB Group, `senseai-eco`.

### Releasing

Toolbox has two types of releases: versioned and continuous. It uses Google
Cloud project, `database-toolbox`.

* **Versioned Release:** Official, supported distributions tagged as `latest`.
  The release process is defined in
  [versioned.release.cloudbuild.yaml](.ci/versioned.release.cloudbuild.yaml).
* **Continuous Release:** Used for early testing of features between official
  releases and for end-to-end testing. The release process is defined in
  [continuous.release.cloudbuild.yaml](.ci/continuous.release.cloudbuild.yaml).
* **GitHub Release:** `.github/release-please.yml` automatically creates GitHub
  Releases and release PRs.

### How-to Release a new Version

1. [Optional] If you want to override the version number, send a
   [PR](https://github.com/googleapis/genai-toolbox/pull/31) to trigger
   [release-please](https://github.com/googleapis/release-please?tab=readme-ov-file#how-do-i-change-the-version-number).
   You can generate a commit with the following line: `git commit -m "chore:
   release 0.1.0" -m "Release-As: 0.1.0" --allow-empty`
1. [Optional] If you want to edit the changelog, send commits to the release PR
1. Approve and merge the PR with the title “[chore(main): release
   x.x.x](https://github.com/googleapis/genai-toolbox/pull/16)”
1. The
   [trigger](https://pantheon.corp.google.com/cloud-build/triggers;region=us-central1/edit/27bd0d21-264a-4446-b2d7-0df4e9915fb3?e=13802955&inv=1&invt=AbhU8A&mods=logs_tg_staging&project=database-toolbox)
   should automatically run when a new tag is pushed. You can view [triggered
   builds here to check the
   status](https://pantheon.corp.google.com/cloud-build/builds;region=us-central1?query=trigger_id%3D%2227bd0d21-264a-4446-b2d7-0df4e9915fb3%22&e=13802955&inv=1&invt=AbhU8A&mods=logs_tg_staging&project=database-toolbox)
1. Update the Github release notes to include the following table:
    1. Run the following command (from the root directory):

        ```
        export VERSION="v0.0.0"
        .ci/generate_release_table.sh
        ```

    1. Copy the table output
    1. In the GitHub UI, navigate to Releases and click the `edit` button.
    1. Paste the table at the bottom of release note and click `Update release`.
1. Post release in internal chat and on Discord.

#### Supported Binaries

The following operating systems and architectures are supported for binary
releases:

* linux/amd64
* darwin/arm64
* darwin/amd64
* windows/amd64

#### Supported Container Images

The following base container images are supported for container image releases:

* distroless

### Automated Tests

Integration and unit tests are automatically triggered via Cloud Build on each
pull request. Integration tests run on merge and nightly.

#### Failure notifications

On-merge and nightly tests that fail have notification setup via Cloud Build
Failure Reporter [GitHub Actions
Workflow](.github/workflows/schedule_reporter.yml).

#### Trigger Setup

Configure a Cloud Build trigger using the UI or `gcloud` with the following
settings:

* **Event:** Pull request
* **Region:** global (for default worker pools)
* **Source:**
  * Generation: 1st gen
  * Repo: googleapis/genai-toolbox (GitHub App)
  * Base branch: `^main$`
* **Comment control:** Required except for owners and collaborators
* **Filters:** Add directory filter
* **Config:** Cloud Build configuration file
  * Location: Repository (add path to file)
* **Service account:** Set for demo service to enable ID token creation for
  authenticated services

### Triggering Tests

Trigger pull request tests for external contributors by:

* **Cloud Build tests:** Comment `/gcbrun`
* **Unit tests:** Add the `tests:run` label

## Repo Setup & Automation

* .github/blunderbuss.yml - Auto-assign issues and PRs from GitHub teams
* .github/renovate.json5 - Tooling for dependency updates. Dependabot is built
  into the GitHub repo for GitHub security warnings
* go/github-issue-mirror - GitHub issues are automatically mirrored into buganizer
* (Suspended) .github/sync-repo-settings.yaml - configure repo settings
* .github/release-please.yml - Creates GitHub releases
* .github/ISSUE_TEMPLATE - templates for GitHub issues
