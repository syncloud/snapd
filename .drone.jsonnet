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
            name: "build",
            image: "golang:1.17-buster",
            commands: [
                "VERSION=$(cat version)",
                "./build.sh $VERSION skip-tests "
            ]
        },
        {
            name: "build release",
            image: "golang:1.17-buster",
            commands: [
                "go test ./syncloud/release",
                "go build -ldflags '-linkmode external -extldflags -static' -o syncloud-release-" + arch + " ./syncloud/release",
                "./syncloud-release-" + arch + " -h"
            ]
        },
        {
            name: "build test apps",
            image: "debian:buster-slim",
            commands: [
              "apt update && apt install -y squashfs-tools",
              "./syncloud/test/testapp1/build.sh",
              "./syncloud/test/testapp2/build.sh"
            ]
        },
        {
            name: "apps.syncloud.org",
            image: "debian:buster-slim",
            detach: true,
            commands: [
              "apt update && apt install nginx tree",
              "mkdir -p log",
              "mkdir -p /var/www/html/releases/stable",
              "mkdir -p /var/www/html/apps",
              "./syncloud-release-" + arch + " -f ./syncloud/test/testapp1/testapp1.snap -b master -t /var/www/html",
              "./syncloud-release-" + arch + " -f ./syncloud/test/testapp2/testapp2.snap -b master -t /var/www/html",
              "cp ./syncloud/test/index-v2 /var/www/html/releases/stable/",
              "tree /var/www/html > log/store.tree.log",
              "systemctl status nginx > log/store.status.log"
            ]
        },
        {
            name: "test",
            image: "debian:buster-slim",
            commands: [
              "VERSION=$(cat version)",
              "./syncloud/test/test.sh $VERSION device"
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
              "pip install syncloud-lib s3cmd",
              "syncloud-upload.sh " + name + " $DRONE_BRANCH $VERSION " + name + "-$VERSION-$(dpkg-architecture -q DEB_HOST_ARCH).tar.gz"
            ]
        },
        {
            name: "artifact",
            image: "appleboy/drone-scp",
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
                    "*.snap",
                    "log/*",
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
                files: "syncloud-release-*",
                overwrite: true,
                file_exists: "overwrite"
            },
            when: {
                event: [ "tag" ]
            }
        },
    ],
    services: [
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
    ]
};

[
    build("arm"),
    build("amd64"),
    build("arm64")
]
