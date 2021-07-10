local name = "snapd";

local build(arch) = {
    kind: "pipeline",
    name: arch,

    platform: {
        os: "linux",
        arch: arch
    },
    workspace: {
        base: "/go",
        path: "src/github.com/snapcore/snapd"
    },
    steps: [
        {
            name: "version",
            image: "debian:buster-slim",
            commands: [
                "echo $(date +%y%m%d)$DRONE_BUILD_NUMBER > version"
            ]
        },
        {
            name: "build",
            image: "golang:1.9.7",
            commands: [
                "VERSION=$(cat version)",
                "./build.sh $VERSION skip-tests "
            ]
        },
        {
            name: "test",
            image: "debian:buster-slim",
            commands: [
              "VERSION=$(cat version)",
              "./test.sh $VERSION device"
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
                    "log/*",
                    "snapd-*.tar.gz"
                ]
            },
            when: {
              status: [ "failure", "success" ]
            }
        }
    ],
    services: [{
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
    }],
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
