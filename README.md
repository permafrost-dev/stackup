<p align="center">
    <img src="assets/stackup-app-512px.png" alt="logo" height="150" style="display: block; height: 150px;">
</p>

# StackUp

---

a single application to spin up your entire dev stack.

## About

`StackUp` is a tool for developers that automates the process of spinning up complicated development environments.  It allows you to defines a series of steps that execute in order on startup and shutdown, as well as a list of server processes that should be started. A good example of a use case for `StackUp` is a web application running a Laravel backend, uses the Horizon queue manager, relies on several containers such as MySQL and Redis, and has a frontend written in Next.js. In this instance a developer would need to spin up all of the containers, run the horizon daemon, start an httpd for the backend, and run `npm run dev` to start the Next.js frontend httpd server. `StackUp` automates this entire process with a single configuration file.

One of the key features of this application is its ability to automate routine tasks. With a simple configuration, you can define a sequence of tasks that your projects require, such as starting containers, running database migrations, or seeding data. This automation not only saves you time but also ensures consistency across your development environment.

It also includes a robust precondition system. Before doing anything, checks can be performed to ensure everything is set up correctly. This feature helps prevent common issues that occur when the environment is not properly configured.

`StackUp` is designed to streamline your development process - it takes care of the repetitive and mundane aspects of managing a development environment, allowing you to focus on what truly matters - writing great code. Whether you're a solo developer or part of a large team, it will significantly enhance your productivity and efficiency.

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

The `tasks` section of the configuration file is used to specify a list of tasks that the application should perform. Tasks are run synchronously in the order they are defined, either on startup or shutdown. 

Each task is defined by a `name`, an optional `if` condition that is a javascript expression that determines if the task should run or be skipped, a `cwd` that can be a javascript expression, an optional `silent` flag, an `on` condition that can be either `startup` or `shutdown`, and a `command`. 

Here is an example of the `tasks` section:

```yaml
tasks:
  - name: start containers
    command: podman-compose up -d
    cwd: '{{ env("LOCAL_BACKEND_PROJECT_PATH") }}'
    on: startup

  - name: run migrations (rebuild db)
    if: hasFlag("seed")
    command: php artisan migrate:fresh --seed
    on: startup

  - name: run migrations (no seeding)
    if: '!hasFlag("seed")'
    command: php artisan migrate
    on: startup

  - name: stop containers
    message: Stopping containers...
    command: podman-compose down
    on: shutdown
```

### Configuration: Servers

The `servers` section of the configuration file is used to specify a list of servers processes that the application should start. Each server is defined by a `name`, a `command`, a `cwd` (current working directory), and an optional `platforms` field. Available `platform` values are `linux`, `darwin` (macOS), or `windows`.

Note that the `cwd` values are wrapped in double braces, which indicates that they should be interpreted as script expressions.

```yaml
servers:
  - name: frontend httpd (linux, macos)
    command: node ./node_modules/.bin/next dev
    cwd: '{{ env("FRONTEND_PROJECT_PATH") }}'
    platforms: ['linux', 'darwin']

  - name: frontend httpd (windows)
    command: npm run dev
    cwd: '{{ env("FRONTEND_PROJECT_PATH") }}'
    platforms: ['windows']

  - name: horizon queue
    command: php artisan horizon
    cwd: '{{ env("LOCAL_BACKEND_PROJECT_PATH") }}'
    platforms: ['linux', 'darwin']

  - name: backend httpd
    command: php artisan serve
    cwd: '{{ env("LOCAL_BACKEND_PROJECT_PATH") }}'
```

### Configuration: Scheduler

The `scheduler` section of the configuration file is used to specify a list of tasks that the application should run on a schedule.
Each entry should contain a `task` id and a `cron` expression.  The `task` value must be equal to the `id` of` a `task` that has been defined and has an `on` value of `schedule`.

Here is an example of the `scheduler` section and its associated `tasks` section:

```yaml
tasks:
  - name: run artisan scheduler
    id: artisan-scheduler
    command: php artisan schedule:run
    cwd: '{{ env("LOCAL_BACKEND_PROJECT_PATH") }}'
    on: schedule

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
