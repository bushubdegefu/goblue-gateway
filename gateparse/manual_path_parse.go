package gateparse

import (
	"encoding/json"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
	"semaygateway.com/gatelogger"
)

// type ServiceSpec map[string]interface{}

func GetTargetLists(service string) ([]string, error) {
	// Read from configuration file
	gate_config, read_error = os.ReadFile("gateway.yaml")

	// alert if file does not exist
	if read_error != nil {
		gatelogger.GateLoggerInfo(read_error.Error())
		return []string{}, read_error
	}

	// parse gateway.yaml config file
	//raw configuration inerface
	var proxy_raw map[string]interface{}

	// unmarshaling using yaml package
	// Unmarshal our input YAML file into empty interface
	if err := yaml.Unmarshal(gate_config, &proxy_raw); err != nil {
		gatelogger.GateLoggerInfo(err.Error())
		return []string{}, err
	}

	// filter map with provided proxy service name to
	// extract service specs from the map
	proxy_raw = proxy_raw["services"].(map[string]interface{})[service].(map[string]interface{})

	// service spec instance
	var instanceConfig ServiceSpec
	// json parse and change to instansiate to servicespec struct
	result, _ := json.Marshal(proxy_raw)
	json.Unmarshal(result, &instanceConfig)
	targets := strings.Split(instanceConfig.Targets, ",")
	return targets, nil
}

func GetRedisTargetLists() ([]string, error) {
	// Read from configuration file
	gate_config, read_error = os.ReadFile("gateway.yaml")

	// alert if file does not exist
	if read_error != nil {
		gatelogger.GateLoggerInfo(read_error.Error())
		return []string{}, read_error
	}

	// parse gateway.yaml config file
	//raw configuration inerface
	var proxy_raw map[string]interface{}

	// unmarshaling using yaml package
	// Unmarshal our input YAML file into empty interface
	if err := yaml.Unmarshal(gate_config, &proxy_raw); err != nil {
		gatelogger.GateLoggerInfo(err.Error())
		return []string{}, err
	}

	// filter map with provided proxy service name to
	// extract service specs from the map
	proxy_raw = proxy_raw["services"].(map[string]interface{})

	redis_list := proxy_raw["redis"].([]interface{})

	// casting []interface{} to []string for target list
	redis_targets := make([]string, len(redis_list))
	for i := range redis_list {
		redis_targets[i] = redis_list[i].(string)
	}

	return redis_targets, nil
}
