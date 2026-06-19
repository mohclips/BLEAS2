module github.com/mohclips/BLEAS2

go 1.17

require (
	github.com/antonfisher/nested-logrus-formatter v1.3.0
	github.com/olivere/elastic/v7 v7.0.22
	github.com/pkg/errors v0.9.1
	github.com/sausheong/ble v0.0.0-20200602153014-61e07e487e3a
	github.com/sirupsen/logrus v1.7.0
	gopkg.in/yaml.v2 v2.2.8
)

require (
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/mgutz/logxi v0.0.0-20161027140823-aebf8a7d67ab // indirect
	github.com/raff/goble v0.0.0-20190909174656-72afc67d6a99 // indirect
	golang.org/x/sys v0.0.0-20191126131656-8a8471f7e56d // indirect
)

replace github.com/sausheong/ble => ./third_party/sausheong-ble
