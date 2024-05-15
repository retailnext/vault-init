package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSecretArnValid(t *testing.T) {
	secretArn := "arn:aws:secretsmanager:us-east-1:418627201366:secret:vault0-ca-private20240430134312612000000003-6F1Z6A"
	name, accountID, region, err := ParseSecretArn(secretArn)
	assert.NoError(t, err)
	assert.Equal(t, "vault0-ca-private20240430134312612000000003", name)
	assert.Equal(t, "418627201366", accountID)
	assert.Equal(t, "us-east-1", region)
}
