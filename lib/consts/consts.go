package consts

const APP_CONFIG_PATH_BASE_NAME = "stackup"
const APP_REPOSITORY = "permafrost-dev/stackup"

const DEFAULT_CWD_SETTING = "{{ getCwd() }}"

var ALL_PLATFORMS = []string{"windows", "linux", "darwin"}
var DEFAULT_ALLOWED_DOMAINS = []string{"raw.githubusercontent.com", "api.github.com"}

const MAX_TASK_RUNS = 99999999

const DEFAULT_CACHE_TTL_MINUTES = 15

var INIT_CONFIG_FILE_CONTENTS string = `name: my stack
description: application stack
version: 1.0.0

settings:
  anonymous-statistics: false
  exit-on-checksum-mismatch: false
  dotenv: ['.env', '.env.local']
  checksum-verification: true
  cache:
    ttl-minutes: 15
  domains:
    allowed:
      - '*.githubusercontent.com'
    hosts:
      - hostname: '*.github.com'
        gateway: allow
        headers:
          - 'Accept: application/vnd.github.v3+json'
  gateway:
    content-types:
      allowed:
        - '*'

includes:
  - url: gh:permafrost-dev/stackup/main/templates/remote-includes/containers.yaml
  - url: gh:permafrost-dev/stackup/main/templates/remote-includes/%s.yaml

# project type preconditions are loaded from included file above
preconditions:

startup:
  - task: start-containers

shutdown:
  - task: stop-containers

servers:

scheduler:

# tasks are loaded from included files above
tasks:
`
