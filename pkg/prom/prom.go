package prom

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/eggfoobar/promdiff/pkg/config"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
)

type QueryResult struct {
	Namespace string
	Target    string
	Value     string
}

type Result struct {
	Query     string
	Name      string
	Changed   QueryResult
	Unchanged QueryResult
}

func FetchData(config config.Config) ([]Result, error) {

	unchangedClient, err := newAPIClient(config.Unchanged.Host, config.Unchanged.Port, config.Unchanged.Token)
	if err != nil {
		return nil, err
	}

	changedClient, err := newAPIClient(config.Changed.Host, config.Changed.Port, config.Changed.Token)
	if err != nil {
		return nil, err
	}

	results := []Result{}
	for _, q := range config.Queries {
		changed, err := query(changedClient, q.Query, config.Changed.Name)
		if err != nil {
			fmt.Printf(" ❌ Query Failed On (%s): (\033[33m%s\033[0m)\n|_ %s\n", config.Changed.Name, q.Name, err)
		}
		unchanged, err := query(unchangedClient, q.Query, config.Unchanged.Name)
		if err != nil {
			fmt.Printf(" ❌ Query Failed On (%s): (\033[33m%s\033[0m)\n|_ %s\n", config.Unchanged.Name, q.Name, err)
		}
		results = append(results, Result{
			Changed:   changed,
			Unchanged: unchanged,
			Query:     q.Query,
			Name:      q.Name,
		})
	}

	return results, nil
}

func query(c v1.API, query, target string) (QueryResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	done := QueryResult{Target: target}

	result, warnings, err := c.Query(ctx, query, time.Now())
	if err != nil {
		return done, fmt.Errorf("error querying Prometheus: %s", err)
	}

	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	switch v := result.(type) {
	case model.Vector:
		for _, r := range v {

			val := map[string]interface{}{}

			bs, err := r.MarshalJSON()
			if err != nil {
				fmt.Println(r.Metric.String())
				fmt.Println(err)
			}
			if err == nil {
				err = json.Unmarshal(bs, &val)
				if err != nil {
					fmt.Println(r.Metric.String())
					fmt.Println(err)
				}
				metric := val["metric"].(map[string]interface{})
				namespace := ""
				if val, ok := metric["namespace"]; ok {
					if namespace, ok = val.(string); !ok {
						namespace = ""
					}
				}
				done.Namespace = namespace
			}

			done.Value = r.Value.String()
			break
		}
	}
	return done, nil
}

func newAPIClient(host, port, secret string) (v1.API, error) {
	if len(host) == 0 {
		host = "https://localhost"
	}
	cfg := api.Config{
		Address: fmt.Sprintf("%s:%s", host, port),
	}

	if len(secret) >= 0 {
		rt := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 5 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		cfg.RoundTripper = promConfig.NewAuthorizationCredentialsRoundTripper("Bearer", promConfig.Secret(secret), rt)
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return v1.NewAPI(client), nil
}
