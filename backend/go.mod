module github.com/mooseburgr/plex-utils

go 1.16

require (
	github.com/GoogleCloudPlatform/functions-framework-go v1.5.2
	github.com/gin-gonic/gin v1.7.7
	github.com/go-co-op/gocron v1.11.0
	github.com/jrudio/go-plex-client v0.0.0-20220106065909-9e1d590b99aa
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.20.0
)

replace github.com/jrudio/go-plex-client v0.0.0-20220106065909-9e1d590b99aa => github.com/mooseburgr/go-plex-client v0.0.0-20220130204429-0b729fc6f3c8

require (
	cloud.google.com/go/functions v1.1.0 // indirect
	github.com/cloudevents/sdk-go/v2 v2.8.0 // indirect
	github.com/go-playground/validator/v10 v10.10.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/ugorji/go v1.2.6 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce // indirect
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
