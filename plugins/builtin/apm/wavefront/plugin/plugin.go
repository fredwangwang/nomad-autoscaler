package plugin

import (
	"fmt"
	"strconv"
	"time"

	wf_api "github.com/WavefrontHQ/go-wavefront-management-api"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad-autoscaler/plugins"
	"github.com/hashicorp/nomad-autoscaler/plugins/apm"
	"github.com/hashicorp/nomad-autoscaler/plugins/base"
	"github.com/hashicorp/nomad-autoscaler/sdk"
)

const (
	pluginName = "wavefront"

	configKeyAddress = "address"
	configKeyAPIKey  = "wf_api_key"
)

var (
	PluginID = plugins.PluginID{
		Name:       pluginName,
		PluginType: sdk.PluginTypeAPM,
	}

	PluginConfig = &plugins.InternalPluginConfig{
		Factory: func(l hclog.Logger) interface{} { return NewWavefrontPlugin(l) },
	}

	pluginInfo = &base.PluginInfo{
		Name:       pluginName,
		PluginType: sdk.PluginTypeAPM,
	}
)

type APMPlugin struct {
	client *wf_api.Client
	config map[string]string
	logger hclog.Logger
}

func NewWavefrontPlugin(log hclog.Logger) apm.APM {
	return &APMPlugin{
		logger: log,
	}
}

func (a *APMPlugin) SetConfig(config map[string]string) error {

	a.config = config

	// If the address is not set, or is empty within the config, any client
	// calls will fail. It seems logical to catch this here rather than just
	// let queries fail.
	addr, ok := a.config[configKeyAddress]
	if !ok || addr == "" {
		return fmt.Errorf("%q config value cannot be empty", configKeyAddress)
	}
	token, ok := a.config[configKeyAPIKey]
	if !ok || token == "" {
		return fmt.Errorf("%q config value cannot be empty", configKeyAPIKey)
	}

	client, err := wf_api.NewClient(
		&wf_api.Config{
			Address: addr,
			Token:   token,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to initialize Wavefront client: %v", err)
	}

	a.client = client

	return nil
}

func (a *APMPlugin) PluginInfo() (*base.PluginInfo, error) {
	return pluginInfo, nil
}

func (a *APMPlugin) Query(q string, r sdk.TimeRange) (sdk.TimestampedMetrics, error) {
	m, err := a.QueryMultiple(q, r)
	if err != nil {
		return nil, err
	}

	switch len(m) {
	case 0:
		return sdk.TimestampedMetrics{}, nil
	case 1:
		return m[0], nil
	default:
		return nil, fmt.Errorf("query returned %d metric streams, only 1 is expected", len(m))
	}
}

func (a *APMPlugin) QueryMultiple(q string, r sdk.TimeRange) ([]sdk.TimestampedMetrics, error) {
	a.logger.Debug("querying Wavefront", "query", q, "range", r)

	query := a.client.NewQuery(
		&wf_api.QueryParams{
			QueryString: q,
			EndTime:     strconv.FormatInt(r.To.Unix(), 10),
			StartTime:   strconv.FormatInt(r.From.Unix(), 10),
			Granularity: "s",
			StrictMode:  true,
		},
	)

	// TODO: execute does not have cancel for timeout.
	result, err := query.Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to query: %v", err)
	}

	if result.Warnings != "" {
		a.logger.Warn("wavefront query returned warning", "warning", result.Warnings)
	}

	if result.ErrType != "" {
		return nil, fmt.Errorf("query returned error: %s: %s", result.ErrType, result.ErrMessage)
	}

	return parseTimeSeries(result.TimeSeries)
}

func parseTimeSeries(tss []wf_api.TimeSeries) ([]sdk.TimestampedMetrics, error) {
	result := make([]sdk.TimestampedMetrics, len(tss))
	for tsi, ts := range tss {
		sdkTs := make(sdk.TimestampedMetrics, len(ts.DataPoints))
		for dpi, dp := range ts.DataPoints {
			sdkTs[dpi].Timestamp = time.Unix(int64(dp[0]), 0)
			sdkTs[dpi].Value = dp[1]
		}
		result[tsi] = sdkTs
	}

	return result, nil
}
