<p align="center">
    <img src="assets/stackup-app-512px.png" alt="logo" height="150" style="display: block; height: 150px;">
</p>

# StackUp

---

Spin up your entire dev stack with one command.  

`StackUp` offers many features and advanced functionality. Here are some of the highlights:

- Define tasks that run on startup, shutdown, or on a schedule.
- Customize tasks and preconditions using javascript.
- Run tasks on a cron schedule, i.e. running `php artisan schedule:run` once every minute.
- Load remote configurations via http or S3.
- Fast in-memory cache skips http requests when possible.
- Http request gateway prevents unwanted access to remote urls, domains and content types.
- Send notifications with Telegram and Slack integrations.

## Table of Contents

- [StackUp](#stackup)
  - [Table of Contents](#table-of-contents)
  - [About](#about)
  - [Running StackUp](#running-stackup)
  - [Configuration](#configuration)
    - [Configuration: Settings](#configuration-settings)
      - [Configuration: Settings: Gateway](#configuration-settings-gateway)
      - [Configuration: Settings: Domains](#configuration-settings-domains)
      - [Configuration: Settings: Notifications](#configuration-settings-notifications)
      - [Configuration: Settings: Notifications: Telegram](#configuration-settings-notifications-telegram)
      - [Configuration: Settings: Notifications: Slack](#configuration-settings-notifications-slack)
    - [Configuration: Environment Variables](#configuration-environment-variables)
    - [Configuration: Includes](#configuration-includes)
    - [Configuration: Preconditions](#configuration-preconditions)
    - [Configuration: Tasks](#configuration-tasks)
    - [Configuration: Startup \& Shutdown](#configuration-startup--shutdown)
    - [Configuration: Servers](#configuration-servers)
    - [Configuration: Scheduler](#configuration-scheduler)
    - [Example Configurations](#example-configurations)
  - [Integrations](#integrations)
    - [Integration: dotenv-vault](#integration-dotenv-vault)
    - [Integration: Telegram Notifications](#integration-telegram-notifications)
    - [Integration: Slack Notifications](#integration-slack-notifications)
    - [Integration: Desktop Notifications](#integration-desktop-notifications)
  - [Scripting](#scripting)
    - [Available Functions](#available-functions)
    - [Script Classes](#script-classes)
      - [`ComposerJson`](#composerjson)
      - [`PackageJson`](#packagejson)
      - [`RequirementsTxt`](#requirementstxt)
      - [`SemVer`](#semver)
    - [Environment Variables](#environment-variables)
  - [Dynamic Tasks](#dynamic-tasks)
    - [Initialization Script](#initialization-script)
  - [Setup](#setup)
  - [Building the project](#building-the-project)
  - [Contributing](#contributing)
  - [Security Vulnerabilities](#security-vulnerabilities)
  - [Credits](#credits)
  - [License](#license)

## About

`StackUp` is a scriptable tool for developers that automates the process of spinning up complicated development environments.  It allows you to defines a series of steps that execute in order on startup and shutdown, as well as a list of server processes that should be started.  Additionally, `StackUp` runs an event loop while the server processes are running, allowing you to run tasks on a cron schedule.

One of the key features of this application is its ability to automate routine tasks. With a simple configuration, you can define a sequence of tasks that your project requires, such as starting containers, running database migrations, or seeding data. This automation not only saves you time but also ensures consistency across your development environment.

It also includes a robust, scriptable precondition system. Before doing anything, checks can be performed to ensure everything is set up correctly. This feature helps prevent common issues that occur when the environment is not properly configured.

## Running StackUp

To run `StackUp`, simply run the binary in a directory containing a `stackup.yaml` or `stackup.dist.yaml` configuration file:

```bash
stackup
```

or, specify a configuration filename:

```bash
stackup --config stackup.dev.yaml
```

To generate a new configuration file to get started, run `init`: 

```bash
stackup init
```

`StackUp` checks if it is running the latest version on startup.  To disable this behavior, use the `--no-update-check` flag:

```bash
stackup --no-update-check
```

## Configuration

The application is configured using a YAML file named `stackup.yaml` and contains five required sections: `preconditions`, `tasks`, `startup`, `shutdown`, and `scheduler`.
There are also optional `settings`, `includes` and `init` sections that can be used to configure and initialize the application.

### Configuration: Settings

The `settings` section of the configuration file is used to configure the application.  The following settings are available:

| field     | description                        | required? |
|-----------|------------------------------------|-----------|
| `anonymous-stats` | `boolean` value specifying whether to send anonymous usage statistics, defaults to `false` | no |
| `defaults.tasks.path` | default path for tasks | no |
| `defaults.tasks.platforms` | default platforms for tasks | no |
| `defaults.tasks.silent` | default silent setting for tasks | no |
| `domains.allowed` | array of domain names that can be accessed (downloads/urls/includes), wildcards are supported. | no |
| `domains.hosts` | array of host settings, such as headers, wildcards are supported. | no |
| `dotenv`  | array of `.env` filenames to load  | no        |
| `cache.ttl-minutes` | number of minutes to cache remote files | no |
| `checksum-verification` | `boolean` value specifying if remote file checksums should be verified, defaults to `true` | no |
| `exit-on-checksum-mismatch` | `boolean` value specifying whether to exit if a checksum mismatch occurs when including a remote file | no |

Example `settings` section:

```yaml
name: my stack
version: 1.0.0

settings:
  anonymous-stats: true # opt-in to sending anonymous usage statistics, default is false.
  dotenv: ['.env', '.env.local'] # loads both `.env` and `.env.local` files, defaults to `.env`.
  exit-on-checksum-mismatch: false # do not exit if a checksum mismatch occurs, defaults to true.
  checksum-verification: false # do not verify checksums, defaults to true.
  cache:
    ttl-minutes: 60 # cache remote files for 60 minutes, defaults to 5 minutes.
  domains:
    allowed:
      # domains allowed for remote file downloads and remote file includes.
      - raw.githubusercontent.com
      - api.github.com
      - app.posthog.com # allow anonymous usage statistics to be sent
    hosts:
    - hostname: api.github.com
      gateway: allow
      headers:
        - 'Authorization: token $GITHUB_TOKEN'
        - 'Accept: application/vnd.github.v3+json'
    - hostname: '*.githubusercontent.com'
      gateway: allow
      headers:
        - 'Authorization: token $GITHUB_TOKEN'
        - 'Accept: application/vnd.github.v3+json'
  defaults:
    tasks:
      silent: true
      path: $LOCAL_BACKEND_PROJECT_PATH
      platforms: ['windows']

tasks:
  - id: task-1
    command: printf "hello world\n"
    # path: defaults to $LOCAL_BACKEND_PROJECT_PATH
    # silent: defaults to true
    # platforms: defaults to ['windows']

  - id: task-2
    command: printf "goodbye world\n"
    path: $FRONTEND_PROJECT_PATH # overrides the default
    platforms: ['linux', 'darwin'] # overrides the default
```

#### Configuration: Settings: Gateway

The `gateway` field of the `settings` section of the configuration file is used to customize the behavior of the http request gateway when accessing remote files or urls.  Both `blocked` and `allowed` content types can be specified as exact strings or wildcards, but both fields are optional.

The default value for `blocked` is to not block any content types, and the default value for `allowed` is to allow all content types.

```yaml
settings:
  gateway:
    content-types:
      blocked:
        - audio/*
        - font/*
        - image/*
        - multipart/*
        - video/*
      allowed:
        - application/json
        - application/vnd.api+json
        - application/vnd.github.*
        - application/xml
        - application/yaml
        - text/*
```

#### Configuration: Settings: Domains

The `domains` section of the configuration file is used to specify a list of domain names that can be accessed when downloading remote files
or including remote files.  Wildcards are supported, such as `*.github.com`.  
If the `domains` section is not specified, default values of `raw.githubusercontent.com` and `api.github.com` are used for the allow list,
and all other domains are blocked.

The `hosts` section of `domains` allows the configuration of headers to send with requests to specific hosts. Hostnames may be fully-qualified
hostnames, or may contain wildcards.  For example, `*.githubusercontent.com` will match `raw.githubusercontent.com` and `gist.githubusercontent.com`.
If the `gateway` field is set to `allow`, the hostname will be added to the list of allowed domains automatically, and defaults to `true`.

> The domain allow list is applied to all url access, including when `StackUp` checks to see if it is running the latest version,
> or when sending anonymous opt-in analytics.

```yaml
domains:
  allowed:
    - raw.githubusercontent.com
    - '*.github.com'
  hosts:
    - hostname: api.github.com
      gateway: allow
      headers:
        - 'Authorization: token $GITHUB_TOKEN'
        - 'Accept: application/vnd.github.v3+json'
    - hostname: '*.githubusercontent.com'
      gateway: allow
      headers:
        - 'Authorization: token $GITHUB_TOKEN'
        - 'Accept: application/vnd.github.v3+json'
```

#### Configuration: Settings: Notifications

StackUp provides the ability to send notifications via several integrations.

```yaml
settings:
  notifications:
    # integration1 settings
    # integration2 settings, etc.
```

#### Configuration: Settings: Notifications: Telegram

To send notifications via Telegram, add a `telegram` section to the `notifications` section of the configuration file.  The `telegram` section should contain an `api-key` and a `chat-ids` field.  The `api-key` field should contain the Telegram bot token, and the `chat-ids` field should be an array of chat ids of the users or groups to send notifications to.  The chat ids may either be a string, number, or an environment variable that contains a chat id.

```yaml
settings:
  notifications:
    telegram:
      api-key: $TELEGRAM_API_KEY
      chat-ids: [$TELEGRAM_CHAT_ID1, $TELEGRAM_CHAT_ID2]
```

For more information on the Telegram integration, see the [Telegram Notifications](#integration-telegram-notifications) section of the [Integrations](#integrations) documentation.

#### Configuration: Settings: Notifications: Slack

To send notifications via Slack, add a `slack` section to the `notifications` section of the configuration file.  The `slack` section should contain `webhook-url` and `channel-ids` fields.  The `webhook-url` field should contain the Slack webhook url to send notifications to, and the `channel-ids` field should be an array of channel names to send notifications to.

```yaml
settings:
  notifications:
    slack:
      webhook-url: $SLACK_WEBHOOK_URL
      channel-ids: [$SLACK_CHANNEL_1, $SLACK_CHANNEL_2]
```

For more information about the Slack integration, see the [Slack Notifications](#integration-slack-notifications) section of the [Integrations](#integrations) documentation.

### Configuration: Environment Variables

Environment variables can be defined in the optional `env` section of the configuration file.  These variables can be referenced in other sections of the configuration file using the `env()` function or by prefixing the variable name with `$` (e.g. `$MY_VAR`).

```yaml
env:
  - MY_ENV_VAR_ONE=test1234
  - MY_ENV_VAR_TWO=1234test
```

### Configuration: Includes

The `includes` section of the configuration file is used to specify a list of filenames, file urls, or s3 urls that should be merged with the configuration.  This is useful for splitting up a large configuration file into smaller, more manageable files or reusing commonly-used tasks, init scripts, or preconditions. Startup, shutdown, servers, and scheduled tasks are not merged from the included files.

Included urls can be prefixed with `gh:` to indicate that the file should be fetched from GitHub.  For example, `gh:permafrost-dev/stackup/main/templates/stackup.dist.yaml` will fetch the `stackup.dist.yaml` file from the `permafrost-dev/stackup` repository on GitHub.
Add a `headers` field to the `url` entry to specify headers to send with the request.  The `headers` field should be an array of strings, where each string is a header to send with the request.  The header value can be a javascript expression if wrapped in double braces.  For example:

```yaml
  - url: gh:permafrost-dev/stackup/main/templates/remote-includes/node.yaml
    headers:
      - 'Authorization: token $GITHUB_TOKEN' 

  - url: gh:permafrost-dev/stackup/main/templates/remote-includes/php.yaml
    headers:
      - '{{ "Authorization: token " + $myGithubTokenVar }}'
```

To import a file from an S3 bucket, prefix the url with `s3:`. For example, `s3:hostname/my-bucket-name/my-config.yaml` will fetch the `my-config.yaml` file from the `my-bucket-name` bucket on `hostname`. Amazon S3 and Minio are supported.

Included files can be specified with either a relative or absolute pathname.  Relative pathnames are relative to the directory containing the configuration file.  Absolute pathnames are relative to the current working directory.

```yaml
includes:
  # include a remote file from github
  - url: gh:permafrost-dev/stackup/main/templates/remote-includes/containers.yaml
    verify: false # optional, defaults to true

  - url: gh:permafrost-dev/stackup/main/templates/remote-includes/node.yaml
    headers:
      # headers to send with the request, javascript must be wrapped in double braces
      - 'Authorization: token $GITHUB_TOKEN'
      - '{{ "X-Some-Header: " + getEnv("GITHUB_TOKEN") }}'

  # include a local file
  - file: python.yaml

  # include a remote file from a minio/s3 bucket
  - url: s3:127.0.0.1:9000/stackup-includes/python.yaml
    access-key: $S3_KEY # access key loaded from `.env` or `env` section
    secret-key: $S3_SECRET # secret key env loaded from `.env` or `env` section
    secure: false # optional, defaults to true
```

If the optional field `verify` is set to `false`, the application will not attempt to verify the checksum of the file before fetching it.  This may be useful for files that are frequently updated, but is not recommended.

If the optional `checksum-url` field is not specified, the application will attempt to fetch the checksum file from the same location as the included file, but with the `.sha256` extension appended to the filename.  For example, if the included file is `gh:permafrost-dev/stackup/main/templates/remote-includes/containers.yaml`, the checksum file will be fetched from `gh:permafrost-dev/stackup/main/templates/remote-includes/containers.yaml.sha256`.

Alternatively, a single `checksums.sha256.txt` or `checksums.sha512.txt` file can exist instead of separate `*.sha256/sha512` files for each file in the `includes` section.  The file should contain a list of checksums for each included file, one per line, in the format `checksum filename` (where `filename` is the base filename only). This is the format generated by the `sha256sum/sha512sum` command-line utilities.  For example:

```text
9e0d9fea90950908c356734df89bfdff4984de4a6143fe32c404cfbc91984fb7  containers.yaml
69b87009a87e38e5470191a9e40c441ce963fb4cf260fd44cf5f032b9566454a  laravel.yaml
```

See the [example configuration](./templates/stackup.dist.yaml) for an example of using includes, the [example remote includes](./templates/remote-includes) for examples of remote include templates, and the [example checksum file](./templates/remote-includes/checksums.sha256.txt) for an example of a checksum file.


Valid algorithms are `sha256` or `sha512`, and checksum files may be generated with the `sha256sum` or `sha512sum` command line utilities.

### Configuration: Preconditions

The `preconditions` section of the configuration file is used to specify a list of conditions that must be met before the tasks and servers can run. Each precondition is defined by a `name` and a `check`. The `name` is a human-readable description of the precondition, and the `check` is a javascript expression that returns a boolean value indicating whether the precondition is met. Unlike other fields, the `check` field does not need to be wrapped in double braces; it is always interpreted as a javascript expression.

Here is an example of the `preconditions` section:

```yaml
preconditions:
    - name: frontend project exists
      check: fs.Exists($FRONTEND_PROJECT_PATH)

    - name: backend project has docker-compose file
      check: fs.Exists($LOCAL_BACKEND_PROJECT_PATH + "/docker-compose.yml")

    - name: backend project is laravel project
      check: fs.Exists($LOCAL_BACKEND_PROJECT_PATH + "/artisan")
```

Preconditions can be configured to run a task or script on failure. If a `on-fail` attribute is specified for a precondition, the application will run the task or script when the precondition fails.  The `on-fail` attribute can be a task `id` or a javascript expression.  If specified, the precondition will re-run in an attempt to successfully pass the check. The maximum number of retries is specified by the `max-retries` attribute, and defaults to 0 (no retries).

```yaml
preconditions:
    - name: check for missing text file
      check: fs.Exists("missing.txt")
      on-fail: '{{ fs.WriteFile("missing.txt", "test") }}'
      max-retries: 1
```

This functionality can be used to configure projects, install dependencies, or perform other tasks that are required for the project to run.  Consider the following:

```yaml
preconditions:
  - name: ensure php dependencies are installed
    check: fs.Exists("vendor") && fs.IsDirectory("vendor")
    on-fail: install-composer-deps
    max-retries: 1

tasks:
  - name: install composer dependencies
    id: install-composer-deps
    command: composer install --no-interaction
```

### Configuration: Tasks

The `tasks` section of the configuration file is used to specify all tasks that can be run during startup, shutdown, as a server, or as a scheduled task.

Items in `tasks` follow this structure:

| field     | description                                                                                                | required? |
|-----------|------------------------------------------------------------------------------------------------------------|-----------|
| `name`      | The name of the task (e.g. `spin up containers`)                                                           | no        |
| `id`        | A unique identifier for the task (e.g. `start-containers`)                                                 | yes       |
| `if`        | A condition that must be true for the task to run (e.g. `hasFlag('seed')`)                                 | no        |
| `command`   | The command to run for the task (e.g. `podman-compose up -d`)                                              | yes       |
| `path`      | The path to the directory where the command should be run `(default: current directory)`. this may be a reference to an environment variable without wrapping it in braces, e.g. `$BACKEND_PROJECT_PATH` | no        |
| `silent`    | Whether to suppress output from the command `(default: false)`                                               | no        |
| `platforms` | A list of platforms where the task should be run `(default: all platforms)`                                  | no        |
| `maxRuns`   | The maximum number of times the task can run (0 means always run) `(default: 0)`                             | no        |

Note that the `command` and `path` values can be wrapped in double braces to be interpreted as a javascript expression.

Here is an example of the `tasks` section:

```yaml
tasks:
  - name: spin up containers
    id: start-containers
    command: podman-compose up -d
    path: $LOCAL_BACKEND_PROJECT_PATH
    silent: true

  - name: run migrations (rebuild db)
    id: run-migrations-fresh
    if: hasFlag("seed")
    command: php artisan migrate:fresh --seed
    path: $LOCAL_BACKEND_PROJECT_PATH

  - name: run migrations (no seeding)
    id: run-migrations-no-seed
    if: '!hasFlag("seed")'
    command: php artisan migrate
    path: '{{ env("LOCAL_BACKEND_PROJECT_PATH") }}'

  - name: frontend httpd (linux, macos)
    id: frontend-httpd-linux
    command: node ./node_modules/.bin/next dev
    path: '{{ env("FRONTEND_PROJECT_PATH") }}'
    platforms: ['linux', 'darwin']
```

However the only required fields are `id` and `command`:

```yaml
tasks:
  - id: start-containers
    command: docker-compose up -d

  - id: stop-containers
    command: docker-compose down
```

### Configuration: Startup & Shutdown

The `startup` and `shutdown` sections of the configuration define the tasks that should be run synchronously during either startup or shutdown.  The values listed must match a defined task `id`.

```yaml
startup:
  - task: start-containers
  - task: run-migrations

shutdown:
  - task: stop-containers
```

### Configuration: Servers

The `servers` section of the configuration file is used to specify a list of tasks that the application should start as server processes. The values listed must match a defined task `id`.

Server processes are started in the order that they are defined, however the application does not wait for the process to start before starting the next task.  If you need to wait for a task to complete, it should be run in the `startup` configuration section.

```yaml
servers:
  - task: frontend-httpd
  - task: backend-httpd
  - task: horizon-queue
```

### Configuration: Scheduler

The `scheduler` section of the configuration file is used to specify a list of tasks that the application should run on a schedule.
Each entry should contain a `task` id and a `cron` expression.  The `task` value must be equal to the `id` of a `task` that has been defined.

Here is an example of the `scheduler` section and its associated `tasks` section:

```yaml
tasks:
  - id: run-artisan-scheduler
    command: php artisan schedule:run

scheduler:
    - task: run-artisan-scheduler
      cron: '* * * * *'
```

### Example Configurations

See the [example configuration](./templates/stackup.dist.yaml) for a more complex example that brings up a Laravel-based backend and a Next.js frontend stack.

Working on a standalone Laravel application? Check out the [example laravel configuration](./templates/stackup.laravel.yaml).

## Integrations

StackUp supports several integrations that provide additional functionality.

### Integration: dotenv-vault

`StackUp` includes an integration for `dotenv-vault` and loading encrypted values from `.env.vault` files (see the [dotenv-vault website](https://vault.dotenv.org)).

To load a `.env.vault` file, add an entry to the `env` section named `dotenv://vault`.  This item will cause the `.env.vault` file to
be loaded into the environment, if it exists.  If it does not exist, no action is taken.

```yaml
env:
  - MY_ENV_VAR_ONE=test1234
  - dotenv://vault # loads .env.vault, if it exists
```

### Integration: Telegram Notifications

`StackUp` includes an integration for sending notifications via Telegram.  To configure the integration, see the [Telegram Notifications](#configuration-settings-notifications-telegram) section of the [Configuration: Settings](#configuration-settings) documentation.

Notifications are sent using javascript:

```js
// send a notification to all configured chat ids
notifications.Telegram().Message("hello from stackup, test 123").Send()

// send a notification to a specific chat id
notifications.Telegram().Message("hello from stackup, test 456").To($TELEGRAM_CHAT_ID_1).Send()
```

### Integration: Slack Notifications

`StackUp` includes an integration for sending messages to Slack channels.  To configure the integration, see the [Slack Notifications](#configuration-settings-notifications-slack) section of the [Configuration: Settings](#configuration-settings) documentation.

Notifications are sent using javascript:

```js
// send a notification to all configured chat ids
notifications.Slack().Message("hello from stackup, test 123").Send()

// send a notification to a specific chat id
notifications.Slack().Message("hello from stackup, test 456").To($SLACK_CHANNEL_1).Send()
```

### Integration: Desktop Notifications

`StackUp` includes an integration for displaying desktop notifications. To configure the integration, see the [Desktop Notifications](#configuration-settings-notifications-desktop) section of the [Configuration: Settings](#configuration-settings) documentation.

Notifications are sent using javascript:

```js
// display a desktop notification
notifications.Desktop().Message("hello from stackup").Send()

// optionally provide a notification title
notifications.Desktop().Message("hello from stackup", "some title").Send()
```

## Scripting

Many of the fields in a `Task` can be defined using javascript. To specify an expression to be evaluated, wrap the content in double braces: `{{ env("HOME") }}`.

### Available Functions


| Function   | Arguments         | Description                                                                 |
|----------- |------------------ |---------------------------------------------------------------------------- |
| `binaryExists()`| `name: string`   | returns true if the specified binary exists in `$PATH`, otherwise false       |
| `env()`      | `name: string`      | returns the string value of environment variable `name                        |
| `exists()`   | `filename: string`  | returns true if `filename` exists, false otherwise                          |
| `fetch()`    | `url: string`       | returns the contents of the url `url` as a string; gateway rules apply      |
| `fetchJson()`| `url: string`       | returns the contents of the url `url` as a JSON object; gateway rules apply |
| `fileContains()`| `filename: string, search: string` | returns true if `filename` contains `search`, false otherwise |
| `getCwd()`   | --                | returns the directory stackup was run from                                  |
| `hasEnv()`   | `name: string`      | returns true if the specified environment variable exists, otherwise false  |
| `hasFlag()`  | `name: string`      | returns true if the flag `name` was specified when running the application  |
| `outputOf()`   | `command: string`   | returns the output of the command `command` with spaces trimmed           |
| `platform()` | --                | returns the operating system, one of `windows`, `linux` or `darwin` (macOS) |
| `script()`   | `filename: string`  | returns the output of the javascript located in `filename`                  |
| `selectTaskWhen()` | `conditional: boolean, trueTaskId: string falseTaskId: string` | returns a `Task` object based on the value of `conditional` |
| `semver()` | `version: string` | returns a `SemVer` object based on the value of `version` |
| `statusMessage()` | `message: string` | prints a status message to stdout, without a trailing new line |
| `task()`     | `taskId: string`    | returns a `Task` object with the id `taskId`                                |
| `app.FailureMessage()` | `message: string` | prints a failure message with an X to stdout with a trailing new line |
| `app.StatusLine()` | `message: string` | prints a status message to stdout, with a trailing new line |
| `app.StatusMessage()` | `message: string` | prints a status message to stdout, without a trailing new line |
| `app.SuccessMessage()` | `message: string` | prints a success message with a checkmark to stdout with a trailing new line |
| `app.WarningMessage()` | `message: string` | prints a warning message to stdout with a trailing new line 
| `app.Version()` | -- | returns the current version of `StackUp` |
| `dev.ComposerJson()`| `filename: string` | returns the contents of a composer.json file (`filename`) as a `ComposerJson` object |
| `dev.PackageJson()` | `filename: string` | returns the contents of a package.json file (`filename`) as a `PackageJson` object |
| `dev.RequirementsTxt()` | `filename: string` | returns a requirements.txt file (`filename`) as a `RequirementsTxt` object |
| `fs.Exists()`| `filename: string`  | returns true if `filename` exists, false otherwise                          |
| `fs.GetFiles()` | `path: string`   | returns a list of files in `path`                                           |
| `fs.IsDirectory()` | `pathname: string` | returns true if `pathname` is a directory, false otherwise                  |
| `fs.IsFile()` | `filename: string`  | returns true if `filename` is a file, false otherwise                       |
| `fs.ReadFile()`| `filename: string`  | returns the contents of `filename` as a string                              |
| `fs.ReadJSON()` | `filename: string` | returns the contents of `filename` as a JSON object                         |
| `fs.WriteFile()`| `filename: string, contents: string` | writes `contents` to `filename` |
| `fs.WriteJSON()` | `filename: string, obj: Object` | writes `obj` to `filename` as a JSON object |
| `vars.Get()` | `name: string` | returns the value of the application variable `name` |
| `vars.Has()` | `name: string` | returns true if the application variable `name` exists, otherwise false |
| `vars.Set()` | `name: string, value: any` | sets an application variable `name` to the value `value` |


### Script Classes

#### `ComposerJson`

The `ComposerJson` class is returned by the `dev.ComposerJson()` function and was designed for working with `composer.json` files.  It has the following methods and attributes:

| Name | Arguments | Description |
|--------|-----------|-------------|
| `.GetDependencies()` | -- | returns an array of dependencies |
| `.HasDependency()` | `name: string` | returns true if the composer.json file has dependency named `name`, otherwise false |
| `.GetDependency()` | `name: string` | returns the dependency named `name`, if it exists |
| `.GetDevDependency()` | `name: string` | returns the dev dependency named `name`, if it exists |

#### `PackageJson`

The `PackageJson` class is returned by the `dev.PackageJson()` function and was designed for working with `package.json` files.  It has the following methods and attributes:

| Name | Arguments | Description |
|--------|-----------|-------------|
| `.GetDependencies()` | -- | returns an array of dependencies |
| `.GetDependency()` | `name: string` | returns the dependency named `name`, if it exists |
| `.GetDevDependency()` | `name: string` | returns the dev dependency named `name`, if it exists |
| `.GetScript()` | `name: string` | returns the script named `name`, if it exists |
| `.HasDependency()` | `name: string` | returns true if the package.json file has dependency named `name`, otherwise false |
| `.HasDevDependency()` | `name: string` | returns true if the package.json file has dev dependency named `name`, otherwise false |
| `.HasScript()` | `name: string` | returns true if the package.json file has a script named `name`, otherwise false |

#### `RequirementsTxt`

The `RequirementsTxt` class is returned by the `dev.RequirementsTxt()` function and was designed for working with `requirements.txt` files.  It has the following methods and attributes:

| Name | Arguments | Description |
|--------|-----------|-------------|
| `.GetDependencies()` | -- | returns an array of dependencies |
| `.HasDependency()` | `name: string` | returns true if the requirements.txt file has dependency named `name`, otherwise false |
| `.GetDependency()` | `name: string` | returns the dependency named `name`, if it exists |

#### `SemVer`

The `SemVer` class is returned by the `semver()` function And is used to parse and compare semantic version strings.  It has the following methods and attributes:

| Name | Arguments | Description |
|--------|-----------|-------------|
| `.Compare()` | `version: string` | returns 1 if `version` is greater than the current version, -1 if `version` is less than the current version, and 0 if they are equal |
| `.GreaterThan()` | `version: string` | returns true if `version` is greater than the current version, otherwise false |
| `.Gte()` | `version: string` | returns true if `version` is greater than or equal to the current version, otherwise false |
| `.LessThan()` | `version: string` | returns true if `version` is less than the current version, otherwise false |
| `.Lte()` | `version: string` | returns true if `version` is less than or equal to the current version, otherwise false |
| `.Equals()` | `version: string` | returns true if `version` is equal to the current version, otherwise false |
| `.Major` | -- | value of the major version number |
| `.Minor` | -- | value of the minor version number |
| `.Patch` | -- | value of the patch version number |
| `.String` | -- | the original version string |

### Environment Variables

Environment variables can be accessed using the `env()` function or referenced directly as variables by prefixing the variable name with `$` (e.g. `$HOME`).

```yaml
preconditions:
    - name: backend project has a docker-compose file
      check: fs.Exists($BACKEND_PROJECT_PATH + "/docker-compose.yml")

tasks:
  - name: horizon queue
    id: horizon-queue
    if: dev.composerJson($BACKEND_PROJECT_PATH).HasDependency("laravel/horizon");
    command: php artisan horizon
    path: '{{ $BACKEND_PROJECT_PATH }}'
    platforms: ['linux', 'darwin']
```

## Dynamic Tasks

You can create dynamic tasks using either the `selectTaskWhen()` or `task()` function:

```yaml
tasks:
  - name: frontend httpd (linux, macos)
    id: httpd-linux
    command: node ./node_modules/.bin/next dev
    path: '{{ $FRONTEND_PROJECT_PATH }}'
    platforms: ['linux', 'darwin']

  - name: frontend httpd (windows)
    id: httpd-win
    command: npm run dev
    path: '{{ $FRONTEND_PROJECT_PATH }}'
    platforms: ['windows']

  - name: '{{ selectTaskWhen(platform() == "windows", "httpd-win", "httpd-linux").Name }}'
    id: frontend-httpd
    command: '{{ selectTaskWhen(platform() == "windows", "httpd-win", "httpd-linux").Command }}'
    path: '{{ selectTaskWhen(platform() == "windows", "httpd-win", "httpd-linux").Path }}'
```

This example defines tasks with different commands for each operating system, then defines a `frontend-httpd` task that dynamically selects the correct one:

```yaml
tasks:
  - name: frontend httpd (linux)
    id: frontend-httpd-linux
    command: node ./node_modules/.bin/next dev
    path: '{{ $FRONTEND_PROJECT_PATH }}'

  - name: frontend httpd (macOS)
    id: frontend-httpd-darwin
    command: node ./node_modules/.bin/next dev
    path: '{{ $FRONTEND_PROJECT_PATH }}'

  - name: frontend httpd (windows)
    id: frontend-httpd-windows
    command: npm run dev
    path: '{{ $FRONTEND_PROJECT_PATH }}'
    
  - name: '{{ task("frontend-httpd-" + platform()).Name }}'
    id: frontend-httpd
    command: '{{ task("frontend-httpd-" + platform()).Command }}'
    path: '{{ task("frontend-httpd-" + platform()).Path }}'
```

### Initialization Script

You may add an `init` section to the configuration file to run javascript before the `preconditions` section executes:

```yaml
name: my stack
version: 1.0.0

init: |
  if (binaryExists("podman-compose")) {
    vars.Set("containerEngineBinary", "podman-compose");
  } else {
    vars.Set("containerEngineBinary", "docker-compose");
  }

  app.SuccessMessage("container engine selected: * " + vars.Get("containerEngineBinary"));
```

## Setup

```bash
go mod tidy
```

## Building the project

`StackUp` uses [task](https://github.com/go-task/task) for running tasks, which is a tool similar to `make`. 

```bash
task build
```

---

## Contributing

Please see [CONTRIBUTING](.github/CONTRIBUTING.md) for details.

## Security Vulnerabilities

Please review [our security policy](../../security/policy) on how to report security vulnerabilities.

## Credits

- [Patrick Organ](https://github.com/patinthehat)
- [All Contributors](../../contributors)

## License

The MIT License (MIT). Please see [License File](LICENSE) for more information.
