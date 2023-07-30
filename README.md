<p align="center">
    <img src="assets/stackup-app-512px.png" alt="logo" height="150" style="display: block; height: 150px;">
</p>

# StackUp

---

A single application to spin up your entire dev stack.

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
| `defaults.tasks.path` | default path for tasks | no |
| `defaults.tasks.silent` | default silent setting for tasks | no |
| `defaults.tasks.platforms` | default platforms for tasks | no |

Example `settings` section:

```yaml
name: my stack
version: 1.0.0

settings:
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

### Configuration: Includes

The `includes` section of the configuration file is used to specify a list of filenames or file urls that should be merged with the configuration.  This is useful for splitting up a large configuration file into smaller, more manageable files or reusing commonly-used tasks, init scripts, or preconditions.

These urls can be prefixed with `gh:` to indicate that the file should be fetched from GitHub.  For example, `gh:permafrost-dev/stackup/main/templates/stackup.dist.yaml` will fetch the `stackup.dist.yaml` file from the `permafrost-dev/stackup` repository on GitHub.

```yaml
includes:
  - url: gh:permafrost-dev/stackup/main/templates/remote-includes/containers.yaml
    verify: false # optional, defaults to true
```

If the optional field `verify` is set to `false`, the application will not attempt to verify the checksum of the file before fetching it.  This may be useful for files that are frequently updated, but is not recommended.

```yaml

An additional file, if specified in the `settings` section, defines an index file that contains a list of file urls and their checksums; if specified, the checksum of each file will be compared to the checksum in the index file before fetching the file.  This is useful for ensuring that the file has not been modified since it was last fetched.

```yaml
settings:
  exit-on-checksum-mismatch: false
  remote-index-url: gh:permafrost-dev/stackup/main/templates/stackup-template-index.yaml
```

Here is an example of a remote index file, `stackup-template-index.yaml`:

```yaml
templates:
  - name: container tasks
    location: https://raw.githubusercontent.com/permafrost-dev/stackup/main/templates/remote-includes/containers.yaml
    checksum: 9e0d9fea90950908c356734df89bfdff4984de4a6143fe32c404cfbc91984fb7
    algorithm: sha256
```

Valid algorithms are `sha256` or `sha512`.

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

## Scripting

Many of the fields in a `Task` can be defined using javascript. To specify an expression to be evaluated, wrap the content in double braces: `{{ env("HOME") }}`.

### Available Functions


| Function   | Arguments         | Description                                                                 |
|----------- |------------------ |---------------------------------------------------------------------------- |
| `binaryExists()`| `name: string`   | returns true if the specified binary exists in `$PATH`, otherwise false       |
| `env()`      | `name: string`      | returns the string value of environment variable `name                        |
| `exists()`   | `filename: string`  | returns true if `filename` exists, false otherwise                          |
| `fetch()`    | `url: string`       | returns the contents of the url `url` as a string                           |
| `fetchJson()`| `url: string`       | returns the contents of the url `url` as a JSON object                      |
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
| `workflow()` | --                | returns a `Workflow` object                                                 |
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

`stackup` uses [task](https://github.com/go-task/task) for running tasks, which is a tool similar to `make`. 

```bash
task build
```

---

## Changelog

Please see [CHANGELOG](CHANGELOG.md) for more information on what has changed recently.

## Contributing

Please see [CONTRIBUTING](.github/CONTRIBUTING.md) for details.

## Security Vulnerabilities

Please review [our security policy](../../security/policy) on how to report security vulnerabilities.

## Credits

- [Patrick Organ](https://github.com/patinthehat)
- [All Contributors](../../contributors)

## License

The MIT License (MIT). Please see [License File](LICENSE) for more information.
