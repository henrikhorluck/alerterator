package api

import (
	"fmt"
	"github.com/nais/alerterator/api/inhibitions"
	"github.com/nais/alerterator/api/receivers"
	routes "github.com/nais/alerterator/api/routes"
	"github.com/nais/alerterator/pkg/apis/alerterator/v1"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	alertmanagerConfigMapName         = "nais-prometheus-prometheus-alertmanager"
	alertmanagerTemplateConfigMapName = "alertmanager-template-config"
	alertmanagerConfigName            = "alertmanager.yml"
)

func getConfig(name string, configMapInterface corev1.ConfigMapInterface) (map[interface{}]interface{}, error) {
	configMap, err := configMapInterface.Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failing while retrieving %s configMap: %s", name, err)
	}

	if configMap.Data == nil {
		return nil, fmt.Errorf("alertmanager is not properly set up, data is empty")
	}

	config := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(configMap.Data[alertmanagerConfigName]), config)
	if err != nil {
		return nil, fmt.Errorf("failed while unmarshling %s: %s", alertmanagerConfigName, err)
	}

	return config, nil
}

func updateConfigMap(config map[interface{}]interface{}, configMapInterface corev1.ConfigMapInterface) error {
	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed while marshaling: %s", err)
	}

	configMap, err := configMapInterface.Get(alertmanagerConfigMapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failing while retrieving %s configMap: %s", alertmanagerConfigMapName, err)
	}

	configMap.Data[alertmanagerConfigName] = string(data)
	_, err = configMapInterface.Update(configMap)
	if err != nil {
		return fmt.Errorf("failed while updating %s: %s", alertmanagerConfigMapName, err)
	}

	return nil
}

func AddOrUpdateAlertmanagerConfigMap(configMapInterface corev1.ConfigMapInterface, alert *v1.Alert) error {
	currentConfig, err := getConfig(alertmanagerConfigMapName, configMapInterface)
	if err != nil {
		return err
	}
	latestConfig, err := getConfig(alertmanagerTemplateConfigMapName, configMapInterface)
	if err != nil {
		return err
	}

	updatedRoutes, err := routes.AddOrUpdateRoute(alert, currentConfig, latestConfig)
	if err != nil {
		return fmt.Errorf("failed while adding/updating routes: %s", err)
	}
	latestConfig["route"] = updatedRoutes

	updatedReceivers, err := receivers.AddOrUpdateReceiver(alert, currentConfig)
	if err != nil {
		return fmt.Errorf("failed while adding/updating receivers: %s", err)
	}
	latestConfig["receivers"] = updatedReceivers

	updatedInhibitRules, err := inhibitions.AddOrUpdateInhibition(alert, currentConfig)
	if err != nil {
		return fmt.Errorf("failed while adding/updating inhibitions: %s", err)
	}
	latestConfig["inhibit_rules"] = updatedInhibitRules

	updateConfigMap(latestConfig, configMapInterface)

	return nil
}

func DeleteRouteAndReceiverFromAlertManagerConfigMap(configMapInterface corev1.ConfigMapInterface, alert *v1.Alert) error {
	config, err := getConfig(alertmanagerConfigMapName, configMapInterface)
	if err != nil {
		return err
	}

	err = routes.DeleteRoute(alert, config)
	if err != nil {
		return fmt.Errorf("failed while deleting route: %s", err)
	}

	err = receivers.DeleteReceiver(alert, config)
	if err != nil {
		return fmt.Errorf("failed while deleting receivers: %s", err)
	}

	err = inhibitions.DeleteInhibition(alert, config)
	if err != nil {
		return fmt.Errorf("failed while deleting receivers: %s", err)
	}

	return updateConfigMap(config, configMapInterface)
}