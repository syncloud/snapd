module github.com/snapcore/snapd

go 1.13

require (
	github.com/coreos/go-systemd v0.0.0-20161114122254-48702e0da86b
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2
	github.com/gorilla/mux v1.7.4-0.20190701202633-d83b6ffe499a
	github.com/jessevdk/go-flags v1.4.1-0.20180927143258-7309ec74f752
	github.com/kr/pretty v0.2.2-0.20200810074440-814ac30b4b18 // indirect
	github.com/mvo5/goconfigparser v0.0.0-20200803085309-72e476556adb
	// if below two libseccomp-golang lines are updated, one must also update packaging/ubuntu-14.04/rules
	github.com/mvo5/libseccomp-golang v0.9.1-0.20180308152521-f4de83b52afb // old trusty builds only
	github.com/ojii/gettext.go v0.0.0-20170120061437-b6dae1d7af8a
	github.com/snapcore/bolt v1.3.2-0.20210908134111-63c8bfcf7af8
	github.com/snapcore/squashfuse v0.0.0-20171220165323-319f6d41a041
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20201002202402-0a1ea396d57c
	golang.org/x/sys v0.0.0-20210908233432-aa78b53d3365 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
	gopkg.in/macaroon.v1 v1.0.0-20150121114231-ab3940c6c165
	gopkg.in/mgo.v2 v2.0.0-20180704144907-a7e2c1d573e1
	gopkg.in/retry.v1 v1.0.3
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
	gopkg.in/tylerb/graceful.v1 v1.2.15
	gopkg.in/yaml.v2 v2.3.0
)
