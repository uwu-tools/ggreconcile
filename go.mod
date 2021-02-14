module github.com/justaugustus/ggreconcile

go 1.12

require (
	cloud.google.com/go v0.75.0
	github.com/n3wscott/cli-base v0.0.0-20200320151736-40d38c556506
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	golang.org/x/oauth2 v0.0.0-20210112200429-01de73cf58bd
	google.golang.org/api v0.36.0
	google.golang.org/genproto v0.0.0-20210111234610-22ae2b108f89
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apimachinery v0.18.8
	k8s.io/release v0.7.0
	k8s.io/test-infra v0.0.0-20191024183346-202cefeb6ff5
)

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
