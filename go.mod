module github.com/justaugustus/ggreconcile

go 1.17

require (
	cloud.google.com/go v0.56.0
	github.com/bmatcuk/doublestar v1.1.1
	github.com/google/go-cmp v0.4.0
	golang.org/x/net v0.7.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	google.golang.org/api v0.20.0
	google.golang.org/genproto v0.0.0-20200429120912-1f37eeb960b2
	gopkg.in/yaml.v3 v3.0.0-20190709130402-674ba3eaed22
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/test-infra v0.0.0-20191024183346-202cefeb6ff5
)

require (
	github.com/clarketm/json v1.13.0 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	go.opencensus.io v0.22.3 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/grpc v1.28.0 // indirect
)

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
