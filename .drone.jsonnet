local name = "snapd";

local build(arch) = {
    kind: "pipeline",
    name: arch,

    platform: {
        os: "linux",
        arch: arch
    },
    steps: [
        {
            name: "version",
            image: "debian:buster-slim",
            commands: [
                "echo $DRONE_BUILD_NUMBER > version"
            ]
        },
        {
            name: "build squashfs",
            image: "debian:buster-slim",
            commands: [
                "./build-squashfs.sh"
            ]
        },
        {
            name: "build snapd",
            image: "golang:1.18",
            commands: [
                "VERSION=$(cat version)",
                "./build.sh $VERSION skip-tests "
            ]
        },
        {
            name: "upload",
            image: "python:3.9-buster",
            environment: {
                AWS_ACCESS_KEY_ID: {
                    from_secret: "AWS_ACCESS_KEY_ID"
                },
                AWS_SECRET_ACCESS_KEY: {
                    from_secret: "AWS_SECRET_ACCESS_KEY"
                }
            },
            commands: [
              "VERSION=$(cat version)",
              "pip install s3cmd",
              "./.syncloud/upload.sh $DRONE_BRANCH $VERSION " + name + "-$VERSION-$(dpkg-architecture -q DEB_HOST_ARCH).tar.gz"
            ],
            when: {
                branch: ["stable", "master"]
            }
        },
        {
            name: "artifact",
            image: "appleboy/drone-scp:1.6.4",
            settings: {
                host: {
                    from_secret: "artifact_host"
                },
                username: "artifact",
                key: {
                    from_secret: "artifact_key"
                },
                timeout: "2m",
                command_timeout: "2m",
                target: "/home/artifact/repo/" + name + "/${DRONE_BUILD_NUMBER}-" + arch,
                source: [
                    "snapd-*.tar.gz"
                ]
            },
            when: {
              status: [ "failure", "success" ]
            }
        },
        {
            name: "publish to github",
            image: "plugins/github-release:1.0.0",
            settings: {
                api_key: {
                    from_secret: "github_token"
                },
                files: "snapd-*.tar.gz",
                overwrite: true,
                file_exists: "overwrite"
            },
            when: {
                event: [ "tag" ]
            }
        },
    ],
    volumes: [
        {
            name: "dbus",
            host: {
                path: "/var/run/dbus"
            }
        },
        {
            name: "dev",
            host: {
                path: "/dev"
            }
        },
        {
            name: "shm",
            temp: {}
        }
    ]
};

local promote() = {
    kind: "pipeline",
    type: "docker",
    name: "promote",
    platform: {
        os: "linux",
        arch: "amd64"
    },
    steps: [
    {
        name: "promote",
        image: "python:3.9-buster",
        environment: {
          AWS_ACCESS_KEY_ID: {
              from_secret: "AWS_ACCESS_KEY_ID"
          },
          AWS_SECRET_ACCESS_KEY: {
              from_secret: "AWS_SECRET_ACCESS_KEY"
          }
        },
        commands: [
          "pip install s3cmd",
          "./.syncloud/promote.sh"
        ]
    }
    ],
    trigger: {
      event: [
        "promote"
      ]
    }
};

[
    build("amd64"),
    build("arm64"),
    build("arm"),
    promote()
]
