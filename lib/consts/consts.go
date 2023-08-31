package consts

const APPLICATION_NAME = "stackup"
const APP_REPOSITORY = "permafrost-dev/stackup"

const APP_CONFIG_PATH_BASE_NAME = "stackup"
const APP_ICON_URL string = "https://raw.githubusercontent.com/" + APP_REPOSITORY + "/main/assets/stackup-app-512px.png"

const DEFAULT_CACHE_TTL_MINUTES = 15
const DEFAULT_CWD_SETTING = "{{ getCwd() }}"
var DEFAULT_GATEWAY_MIDDLEWARE = []string{"validateUrl", "verifyFileType", "validateContentType"}

const MAX_TASK_RUNS = 99999999

var ALL_PLATFORMS = []string{"windows", "linux", "darwin"}

var DEFAULT_ALLOWED_DOMAINS = []string{"raw.githubusercontent.com", "api.github.com"}
var DISPLAY_URLS_REMOVABLE = []string{"https://", "github.com", "raw.githubusercontent.com", "s3:"}

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
