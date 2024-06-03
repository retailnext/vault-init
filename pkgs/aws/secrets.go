package aws

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type AWSSM struct {
	ctx    context.Context
	client *secretsmanager.Client
}

const (
	SecretPathExp = `^arn:.*:secretsmanager:(?P<region>[^:]+):(?P<account>[^:]+):secret:(?P<name>[^:]+)-[^-]+$`
)

func NewSecretClient(ctx context.Context, region string) (*AWSSM, error) {
	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return &AWSSM{
		ctx:    ctx,
		client: secretsmanager.NewFromConfig(config),
	}, nil
}

func (s *AWSSM) GetVersion(secretName string, version string) (secretContent []byte, err error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String(version), // VersionStage defaults to AWSCURRENT if unspecified
	}
	result, err := s.client.GetSecretValue(s.ctx, input)
	if err != nil {
		return secretContent, fmt.Errorf("get_aws_secret_fail: %v", err)
	}
	fmt.Println(result.SecretString)
	secretContent = []byte(*result.SecretString)
	return
}

func (s *AWSSM) GetValue(secretName string) ([]byte, error) {
	return s.GetVersion(secretName, "AWSCURRENT")
}

func (s *AWSSM) AddVersion(secretName string, secretContent []byte) (string, error) {
	secretString := string(secretContent)
	addVersionReq := &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(secretName),
		SecretString: &secretString,
	}
	result, err := s.client.PutSecretValue(s.ctx, addVersionReq)
	if err != nil {
		return "", fmt.Errorf("add_aws_secret_fail: %v", err)
	}
	return *result.VersionId, nil
}

func CheckSecretPath(secretPath string) bool {
	match, err := regexp.MatchString(SecretPathExp, secretPath)
	if err != nil {
		slog.Error("error_regexp_secret_path", "error", err)
	}
	return match
}

func ParseSecretArn(secretPath string) (name, accountID, region string, err error) {
	if !CheckSecretPath(secretPath) {
		err = fmt.Errorf("%s fails to match %s", secretPath, SecretPathExp)
		return
	}
	pattern := regexp.MustCompile(SecretPathExp)
	matches := pattern.FindStringSubmatch(secretPath)
	region = matches[1]
	accountID = matches[2]
	name = matches[3]
	return
}
