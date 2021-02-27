package main

import (
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad-autoscaler/plugins"
	wavefront "github.com/hashicorp/nomad-autoscaler/plugins/builtin/apm/wavefront/plugin"
)

func main() {
	plugins.Serve(factory)
}

// factory returns a new instance of the Wavefront APM plugin.
func factory(log hclog.Logger) interface{} {
	return wavefront.NewWavefrontPlugin(log)
}
