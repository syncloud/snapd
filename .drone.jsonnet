local name = "snapd";
local go = "1.18.10";
local debian = 'bookworm-slim';
local python = '3.12-slim-bookworm';
local bootstrap = '25.02';

local build(arch, deb_arch) = {
    kind: "pipeline",
    name: arch,

    platform: {
        os: "linux",
        arch: arch
    },
    steps: [
        {
            name: "version",
            image: "debian:" + debian,
            commands: [
                "echo $DRONE_BUILD_NUMBER > version"
            ]
        },
        {
            name: "build squashfs",
            image: "debian:12-slim",
            commands: [
                "./squashfs-build.sh"
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
                "./squashfs-test.sh"
            ]
        },
        {
            name: "test squashfs d10",
            image: "debian:buster-slim",
            commands: [
                "./squashfs-test.sh"
            ]
        },        {
            name: "test squashfs d12",
            image: "debian:12-slim",
            commands: [
                "./squashfs-test.sh"
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
            image: "python:" + python,
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
              "./.syncloud/upload.sh $DRONE_BRANCH $VERSION " + name + "-$VERSION-" + deb_arch+ ".tar.gz"
            ],
            when: {
                branch: ["stable", "master"],
                event: ["push"]
            }
        },
        {
            name: "promote",
            image: "python:" + python,
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
            ],
            when: {
                branch: ["stable"],
                event: ["push"]
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
              status: [ "failure", "success" ],
              event: [ "push" ]
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
            image: "syncloud/bootstrap-bookworm-" + arch + ":" + bootstrap,
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
            image: "syncloud/bootstrap-bookworm-" + arch + ":" + bootstrap,
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
            image: "syncloud/bootstrap-bookworm-" + arch + ":" + bootstrap,
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
        "tag"
      ]
    },
};

[
    build("amd64", "amd64"),
    build("arm64", "arm64"),
    build("arm", "armhf")
]
