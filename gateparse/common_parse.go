package gateparse

import (
	"encoding/json"
	"os"

	"gopkg.in/yaml.v3"
	"semaygateway.com/gatelogger"
)

var (
	gate_config []byte
	read_error  error
)

type LoadBalanceConfig struct {
	//possible values are  "round_robbin" ,"least_connection"
	Option string `json:"option"`
}

type ListString []string
type RateLimitConfig struct {
	// # possible values for rate limiting
	// #  None
	// # "Token bucket"
	// # "Sliding window counter"
	Source   string `json:"source"`
	Option   string `json:"option"`
	Redis    int    `json:"redis"`
	Limit    int    `json:"limit"`
	Interval int    `json:"interval"`
	Rabbit   int    `json:"rabbit"`
}

type ServiceSpec struct {
	Dicovery bool   `json:"discovery"`
	Targets  string `json:"targets"`
	Socket   string `json:"socket"`
	Prefix   string `json:"prefix"`
}

func GetLoadBalanceMethod(service string) (LoadBalanceConfig, error) {

	gate_config, read_error = os.ReadFile("gateway.yaml")
	if read_error != nil {
		gatelogger.GateLoggerInfo(read_error.Error())
		return LoadBalanceConfig{}, read_error
	}

	//raw configuration inerface
	var proxy_raw map[string]interface{}

	// unmarshaling using yaml package
	// Unmarshal our input YAML file into empty interface
	if err := yaml.Unmarshal(gate_config, &proxy_raw); err != nil {
		gatelogger.GateLoggerInfo(err.Error())
		return LoadBalanceConfig{}, err
	}
	// Getting loadbalnce value from map
	proxy_raw = proxy_raw["services"].(map[string]interface{})[service].(map[string]interface{})["loadbalance"].(map[string]interface{})

	// instantiating loadbalancing config for service
	var instanceConfig LoadBalanceConfig

	//
	result, _ := json.Marshal(proxy_raw)
	// balanceConfig := proxy_raw["service"].(map[string]interface{})["loadbalance"].(map[string]string)
	json.Unmarshal(result, &instanceConfig)

	return instanceConfig, nil
}

func GetRateLimitConfig(service string) (RateLimitConfig, error) {
	// Reading configuration file
	gate_config, read_error = os.ReadFile("gateway.yaml")
	if read_error != nil {
		gatelogger.GateLoggerInfo(read_error.Error())
		return RateLimitConfig{}, read_error
	}

	//raw configuration inerface
	var proxy_raw map[string]interface{}

	// unmarshaling using yaml package
	// Unmarshal our input YAML file into empty interface
	if err := yaml.Unmarshal(gate_config, &proxy_raw); err != nil {
		gatelogger.GateLoggerInfo(err.Error())
		return RateLimitConfig{}, err
	}

	proxy_raw = proxy_raw["services"].(map[string]interface{})[service].(map[string]interface{})["ratelimit"].(map[string]interface{})
	var instanceConfig RateLimitConfig

	// json parse raw map
	result, _ := json.Marshal(proxy_raw)
	json.Unmarshal(result, &instanceConfig)

	return instanceConfig, nil
}

// Returns number of services defiend under the gateway.yaml config file
func GetServiceLists() ([]string, error) {
	//Reading Config list
	gate_config, read_error = os.ReadFile("gateway.yaml")

	if read_error != nil {
		gatelogger.GateLoggerInfo(read_error.Error())
		return []string{}, read_error
	}

	//raw configuration inerface
	var proxy_raw map[string]interface{}

	// unmarshaling using yaml package
	// Unmarshal our input YAML file into empty interface
	if err := yaml.Unmarshal(gate_config, &proxy_raw); err != nil {
		gatelogger.GateLoggerInfo(err.Error())
		return []string{}, err
	}

	// extract service specs from the map
	proxy_raw = proxy_raw["services"].(map[string]interface{})
	service_list := make([]string, 0)

	// making list of services from yaml config
	for keys := range proxy_raw {
		service_list = append(service_list, keys)
	}

	return service_list, nil
}

// Get service prefix other than redis or rabbit
func GetServicePrefix(service string) string {
	// only proceed if service name is not rabbit or redis
	// this are service names used as que by the gateway as token bucket and rate conunter store
	if !(service == "redis") && !(service == "rabbit") {

		gate_config, read_error = os.ReadFile("gateway.yaml")
		if read_error != nil {
			gatelogger.GateLoggerInfo(read_error.Error())
			return read_error.Error()
		}

		//raw configuration inerface
		var proxy_raw map[string]interface{}

		// unmarshaling using yaml package
		// Unmarshal our input YAML file into empty interface
		if err := yaml.Unmarshal(gate_config, &proxy_raw); err != nil {
			gatelogger.GateLoggerInfo(err.Error())
			return err.Error()
		}

		prefix := proxy_raw["services"].(map[string]interface{})[service].(map[string]interface{})["prefix"].(string)
		return prefix
	}
	return "This services do not have a prefix"
}
