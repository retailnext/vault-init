package gcp

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type GCPSSM struct {
	ctx    context.Context
	client *secretmanager.Client
}

const (
	SecretPathExp = "projects\\/([^\\/]+)\\/secrets\\/([^\\/]+)"
)

func NewSecretClient(ctx context.Context) (*GCPSSM, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &GCPSSM{
		ctx:    ctx,
		client: client,
	}, nil
}

func (s *GCPSSM) GetVersion(secretPath string, version string) (secretContent []byte, err error) {
	if !CheckSecretPath(secretPath) {
		return secretContent, fmt.Errorf("wrong secret path: %s does not match %s", secretPath, SecretPathExp)
	}

	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("%s/versions/%s", secretPath, version),
	}
	result, err := s.client.AccessSecretVersion(s.ctx, accessRequest)
	if err != nil {
		return secretContent, fmt.Errorf("get_version_fail: %v", err)
	}
	return result.Payload.Data, nil
}

func (s *GCPSSM) GetValue(secretPath string) ([]byte, error) {
	return s.GetVersion(secretPath, "latest")
}

func (s *GCPSSM) AddVersion(secretPath string, secretContent []byte) (string, error) {
	if !CheckSecretPath(secretPath) {
		return "", fmt.Errorf("wrong secret path: %s does not match %s", secretPath, SecretPathExp)
	}

	addSecretVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secretPath,
		Payload: &secretmanagerpb.SecretPayload{
			Data: secretContent,
		},
	}
	version, err := s.client.AddSecretVersion(s.ctx, addSecretVersionReq)
	if err != nil {
		return "", fmt.Errorf("add_version_fail: %v", err)
	}
	return version.Name, nil
}

func CheckSecretPath(secretPath string) bool {
	match, err := regexp.MatchString(SecretPathExp, secretPath)
	if err != nil {
		slog.Error("error_regexp_secret_path", "error", err)
	}
	return match
}
