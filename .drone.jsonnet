local name = "snapd";
local go = "1.18.10";

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
            image: "debian:9-slim",
            commands: [
                "./build-squashfs.sh"
            ]
        },
        {
            name: "build test store",
            image: "golang:" + go,
            commands: [
              "apt update && apt install -y squashfs-tools",
              ".syncloud/test/build.sh",
              ".syncloud/test/build-apps.sh",
              ".syncloud/test/publish.sh"
            ]
        },
       {
            name: "test squashfs d8",
            image: "debian:jessie-slim",
            commands: [
                "./build/snapd/bin/unsquashfs -ll .syncloud/test/testapp1_1_*.snap"
            ]
        },
        {
            name: "test squashfs d10",
            image: "debian:buster-slim",
            commands: [
                "./build/snapd/bin/unsquashfs -ll .syncloud/test/testapp1_1_*.snap"
            ]
        },        {
            name: "test squashfs d12",
            image: "debian:12-slim",
            commands: [
                "./build/snapd/bin/unsquashfs -ll .syncloud/test/testapp1_1_*.snap"
            ]
        },
        {
            name: "build snapd",
            image: "golang:" + go,
            commands: [
                "VERSION=$(cat version)",
                "./build.sh $VERSION skip-tests "
            ]
        },
        {
            name: "test",
            image: "golang:" + go,
            commands: [
              ".syncloud/test/test.sh"
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
                    "snapd-*.tar.gz",
                    "artifacts/*"
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
    services:
    [
        {
            name: "device",
            image: "syncloud/bootstrap-buster-" + arch,
            privileged: true,
            volumes: [
                {
                    name: "dbus",
                    path: "/var/run/dbus"
                },
                {
                    name: "dev",
                    path: "/dev"
                }
            ]
        },
         {
            name: "api.store.test",
            image: "syncloud/bootstrap-buster-" + arch,
            privileged: true,
            volumes: [
                {
                    name: "dbus",
                    path: "/var/run/dbus"
                },
                {
                    name: "dev",
                    path: "/dev"
                }
            ]
        },
        {
            name: "apps.syncloud.org",
            image: "syncloud/bootstrap-buster-" + arch,
            privileged: true,
            volumes: [
                {
                    name: "dbus",
                    path: "/var/run/dbus"
                },
                {
                    name: "dev",
                    path: "/dev"
                }
            ]
        }
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
    ],
    trigger: {
      event: [
        "push",
        "pull_request",
        "tag"
      ]
    },
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
