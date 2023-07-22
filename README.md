<p align="center">
    <img src="assets/stackup-app-512px.png" alt="logo" height="150" style="display: block; height: 150px;">
</p>

# StackUp

---

a single application to spin up your entire dev stack.

## About

`StackUp` is a tool for developers that automates the process of spinning up complicated development environments.  It allows you to defines a series of steps that execute in order on startup and shutdown, as well as a list of server processes that should be started.

One of the key features of this application is its ability to automate routine tasks. With a simple configuration, you can define a sequence of tasks that your project requires, such as starting containers, running database migrations, or seeding data. This automation not only saves you time but also ensures consistency across your development environment.

It also includes a robust precondition system. Before doing anything, checks can be performed to ensure everything is set up correctly. This feature helps prevent common issues that occur when the environment is not properly configured.

## Configuration

The application is configured using a YAML file. This file contains a list of tasks that the application should perform, as well as a list of servers that the application should start. The file also contains a list of preconditions that must be met before the application can run.

### Configuration: Preconditions

The `preconditions` section of the configuration file is used to specify a list of conditions that must be met before the tasks and servers can run. Each precondition is defined by a `name` and a `check`. The `name` is a human-readable description of the precondition, and the `check` is a javascript expression that returns a boolean value indicating whether the precondition is met. Unlike other fields, the `check` field does not need to be wrapped in double braces; it is always interpreted as a javascript expression.

Here is an example of the `preconditions` section:

```yaml
preconditions:
    - name: frontend project exists
      check: exists(env("FRONTEND_PROJECT_PATH"))

    - name: backend project has docker-compose file
      check: exists(env("LOCAL_BACKEND_PROJECT_PATH") + "/docker-compose.yml")

    - name: backend project is laravel project
      check: exists(env("LOCAL_BACKEND_PROJECT_PATH") + "/artisan")
```

### Configuration: Tasks

The `tasks` section of the configuration file is used to specify all tasks that can be run during startup, shutdown, as a server, or as a scheduled task.

Items in `tasks` follow this structure:

| field     | description                                                                                                | required? |
|-----------|------------------------------------------------------------------------------------------------------------|-----------|
| name      | The name of the task (e.g. `spin up containers`)                                                           | yes       |
| id        | A unique identifier for the task (e.g. `start-containers`)                                                 | yes       |
| if        | A condition that must be true for the task to run (e.g. `hasFlag('seed')`)                                 | no        |
| command   | The command to run for the task (e.g. `podman-compose up -d`)                                              | yes       |
| path      | The path to the directory where the command should be run                                                  | yes       |
| silent    | Whether to suppress output from the command `(default: false)`                                               | no        |
| platforms | A list of platforms where the task should be run `(default: all platforms)`                                  | no        |

Note that the `path` value can be wrapped in double braces to indicate that it should be interpreted as a javascript expression.

Here is an example of the `tasks` section:

```yaml
tasks:
  - name: spin up containers
    id: start-containers
    command: podman-compose up -d
    path: '{{ env("LOCAL_BACKEND_PROJECT_PATH") }}'
    silent: true

  - name: run migrations (rebuild db)
    id: run-migrations-fresh
    if: hasFlag("seed")
    command: php artisan migrate:fresh --seed
    path: '{{ env("LOCAL_BACKEND_PROJECT_PATH") }}'

  - name: run migrations (no seeding)
    id: run-migrations-no-seed
    if: '!hasFlag("seed")'
    run: php artisan migrate
    path: '{{ env("LOCAL_BACKEND_PROJECT_PATH") }}'

  - name: frontend httpd (linux, macos)
    id: frontend-httpd-linux
    command: node ./node_modules/.bin/next dev
    path: '{{ env("FRONTEND_PROJECT_PATH") }}'
    platforms: ['linux', 'darwin']
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
  - name: run artisan scheduler
    id: artisan-scheduler
    command: php artisan schedule:run
    path: '{{ env("LOCAL_BACKEND_PROJECT_PATH") }}'

scheduler:
    - task: artisan-scheduler
      cron: '* * * * *'
```

### Example Configuration

See the [example configuration](./templates/stackup.dist.yaml) for an example that brings up a Laravel-based backend and a Next.js frontend stack.

## Available Functions

Many of the configuration fields can be defined using a javascript expression syntax.
To specify an expression to be evaluated, wrap the content in double braces: `{{ env("HOME") }}`.

| Function   | Arguments         | Description                                                                 |
|----------- |------------------ |---------------------------------------------------------------------------- |
| `binaryExists()`| name: string   | returns true if the specified binary exists in `$PATH`, otherwise false       |
| `env()`      | name: string      | returns the string value of environment variable `name                        |
| `exists()`   | filename: string  | returns true if `filename` exists, false otherwise                          |
| `hasFlag()`  | name: string      | returns true if the flag `name` was specified when running the application  |

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
