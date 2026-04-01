package vault

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/hashicorp/vault/api"
)

func (v *Vault) CreateKVMountIfNotExist(mountPath string) (err error) {
	mountPath = strings.TrimSpace(mountPath)
	if len(mountPath) == 0 {
		return fmt.Errorf("mount path cannot be empty")
	}

	mounts, err := v.client.Sys().ListMounts()
	if err != nil {
		return err
	}
	if mountPath[len(mountPath)-1] != '/' {
		mountPath = mountPath + "/"
	}

	if _, ok := mounts[mountPath]; ok {
		slog.Info("mount already exists")
		return nil
	}

	return v.client.Sys().Mount(mountPath, &api.MountInput{
		Type: "kv",
		Options: map[string]string{
			"version": "2",
		},
	})
}

func (v *Vault) ReadSecret(mountPath, secretPath string) (string, error) {
	apiPath := fmt.Sprintf("%s/data/%s", strings.TrimRight(mountPath, "/"), secretPath)
	secret, err := v.client.Logical().Read(apiPath)
	if err != nil {
		return "", err
	}
	if secret == nil {
		return "", fmt.Errorf("secret not found: %s/%s", mountPath, secretPath)
	}
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected secret data format")
	}
	value, ok := data["value"].(string)
	if !ok {
		return "", fmt.Errorf("value field missing or not a string")
	}
	return value, nil
}

func (v *Vault) WriteSecret(mountPath, secretPath, secretValue string) (err error) {
	apiPath := fmt.Sprintf("%s/data/%s", strings.TrimRight(mountPath, "/"), secretPath)
	_, err = v.client.Logical().Write(apiPath, map[string]interface{}{
		"data": map[string]interface{}{
			"value": secretValue,
		},
	})
	return err
}
