# -*- mode: Python -*-

#
# Command line args
#

# Specifies a list of args like `tilt up foo bar blabla`
config.define_string_list("to-run", args=True)
cfg = config.parse()


TILT_MODE_DEFAULT = "default"

# minimal subset of resources to develop
default_resources = [
  "telegramsearch",
  "telegramsearch-pg",

]

groups = {
    TILT_MODE_DEFAULT: default_resources,
}

# Resources to run, empty array means everything
resources = []

# Run TILT_MODE_DEFAULT by default
tilt_args = cfg.get("to-run", [TILT_MODE_DEFAULT])

for arg in tilt_args:
    if arg in groups:
        resources += groups[arg]
    else:
        # Also support specifying individual services instead of groups,
        # e.g. `tilt up a b d`
        resources.append(arg)


# Tells Tilt to only run specified resources
config.set_enabled_resources(resources)

#
# Plugins
#

load("ext://uibutton", "cmd_button")

#
# Docker compose
#

docker_compose_files = [
    "./docker-compose.yml",
    "./docker-compose.persistent.yml",
]
docker_compose(docker_compose_files)

#
# Labels
#

dc_resource(
    "telegramsearch",
    labels=["bot"],
    resource_deps=[
      "telegramsearch-pg",
    ],
)

dc_resource(
    "telegramsearch-pg",
    labels=["bot"],
)

local_resource(
    "migrate telegramsearch-pg ",
    cmd="DOCKER_BUILDKIT=1 docker build . -f app/telegramsearch/telegramsearch-db/telegramsearch-db.Dockerfile -t yanakipre/telegramsearch-migrations:local && docker run --network yanakipre_net --env DATABASE_URL='postgres://postgres:password@10.30.41.52:5432/telegramsearch' yanakipre/telegramsearch-migrations:local",
    labels=["bot"],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    resource_deps=[
      "telegramsearch-pg",
    ],
)

docker_build(
    "yanakipre/telegramsearch:local",
    context=".",
    dockerfile="app/telegramsearch/cmd/telegramsearch/telegramsearch.Dockerfile",
    only=[
        "app/telegramsearch",
        "internal",
        "go.mod",
        "go.sum",
    ],
)

