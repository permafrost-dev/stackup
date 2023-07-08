<p align="center">
    <img src="assets/stackup-app-512px.png" alt="logo" height="150" style="display: block; height: 150px;">
</p>

# StackUp

---

a single application to manage your entire dev stack.

## About Stackup

The application we've developed is a comprehensive tool designed to manage your entire development stack. It's a one-stop solution that brings together all the elements of your development environment, providing a unified interface to control and monitor each component. This includes everything from your frontend and backend projects, to databases, servers, queues, and even third-party services.

One of the key features of this application is its ability to automate routine tasks. With a simple configuration, you can define a sequence of tasks that your projects require, such as starting containers, running database migrations, or seeding data. This automation not only saves you time but also ensures consistency across your development environment.

It also includes a robust precondition system. Before running tasks or starting servers, checks are performed to ensure everything is set up correctly. This feature helps prevent common issues that can occur when the environment is not properly configured.

In essence, this application is designed to streamline your development process. It takes care of the repetitive and mundane aspects of setting up and managing a development environment, allowing you to focus on what truly matters - writing great code. Whether you're a solo developer or part of a large team, this application can significantly enhance your productivity and efficiency.

## Configuration

The application is configured using a YAML file. This file contains a list of tasks that the application should perform, as well as a list of servers that the application should start. The file also contains a list of preconditions that must be met before the application can run.

### Configuration: Preconditions

The `preconditions` section of the configuration file is used to specify a list of conditions that must be met before the application can run. Each precondition is defined by a `name` and a `check`. The `name` is a human-readable description of the precondition, and the `check` is a function that returns a boolean value indicating whether the precondition is met.

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

### Configuration: Commands

The `commands` section of the configuration file is used to specify a list of custom commands that the application can execute. Each command is defined by a `name`, a `description`, a `command`, an optional `silent` flag, and an optional `on` field.

Here is an example of the `commands` section:

```yaml
commands:
  - name: stop-containers
    description: stop containers
    command: podman-compose down
    silent: true
    on: shutdown
```

In this example, the application defines a custom command:

- Stop containers: The `command` runs the `podman-compose down` command, which stops the containers. The `description` field provides a brief explanation of what the command does. The `silent` flag is set to true, which means that the command's output will not be displayed. The `on` field is set to shutdown, which means that this command will be executed when the application is shutting down.


The `commands` section allows you to extend the functionality of the application by defining your own commands. These commands can be tied to specific events (like startup or shutdown). This provides flexibility in controlling the behavior of your application.

### Configuration: Tasks

The `tasks` section of the configuration file is used to specify a list of tasks that the application should perform. Each task is defined by a `name`, an optional `message`, an optional `if` condition, and a `command`.

Here is an example of the `tasks` section:

```yaml
tasks:
  - name: start containers
    message: Starting containers...
    command: podman-compose up -d

  - name: run migrations (rebuild db)
    if: hasFlag("seed")
    command: php artisan migrate:fresh --seed

  - name: run migrations (no seeding)
    if: hasFlag("seed")
    command: php artisan migrate

  - name: seed temp quickbooks refresh token for dev
    if: true
    command: php artisan qb:create-test-token
```

### Configuration: Servers

The `servers` section of the configuration file is used to specify a list of servers processes that the application should start. Each server is defined by a name, a command, a cwd (current working directory), and an optional platforms field.

Note that the `command` values are wrapped in double braces, which indicates that they should be interpreted as script expressions.

```yaml
servers:
  - name: frontend-httpd
    command: node ./node_modules/.bin/next dev
    cwd: '{{ env("FRONTEND_PROJECT_PATH") }}'

  - name: horizon queue
    command: php artisan horizon
    cwd: '{{ env("LOCAL_BACKEND_PROJECT_PATH") }}'
    platforms: ['linux', 'darwin']

  - name: httpd
    command: php artisan serve
    cwd: '{{ env("LOCAL_BACKEND_PROJECT_PATH") }}'
```

### Configuration: Scheduler

The `scheduler` section of the configuration file is used to specify a list of tasks that the application should run on a schedule, separate from any event loop tasks. 
Each scheduled task is defined by a `name`, a `command`, and a `cron` string.

Here is an example of the `scheduler` section:

```yaml
scheduler:
    - name: say hello every 1 minute
      command: printf "hello world\n"
      cron: '0 */1 * * * *'

    - name: say goodbye every 30 seconds
      command: printf "goodbye\n"
      cron: '*/30 * * * * *'
```

Note that these cron schedules differ from the standard in that you must specify seconds as the first item, followed by the usual items (minute, hour, etc.).

## Available Functions

Many of the configuration fields can be defined using a javascript expression syntax.
To specify an expression to be evaluated, wrap the content in double braces: `{{ myfunc() }}`.

| Function  	| Arguments        	| Description                                                                	|
|-----------	|------------------	|----------------------------------------------------------------------------	|
| env()     	| name: string     	| returns the value environment variable `name`                              	|
| exists()  	| filename: string 	| returns true if `filename` exists, false otherwise                         	|
| hasFlag() 	| name: string     	| returns true if the flag `name` was specified when running the application 	|

## Setup

```bash
go mod tidy
```

## Building the project

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
