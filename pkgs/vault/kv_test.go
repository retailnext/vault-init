package vault

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateKVMountIfNotExist(t *testing.T) {
	ctx := context.Background()
	_, vclient, err := StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = vclient.CreateKVMountIfNotExist("secret")
	assert.NoError(t, err)

	mounts, err := vclient.client.Sys().ListMounts()
	assert.NoError(t, err)
	assert.Contains(t, mounts, "secret/")
	assert.Equal(t, "kv", mounts["secret/"].Type)
	assert.Equal(t, "2", mounts["secret/"].Options["version"])
}

func TestCreateKVMountIfNotExist_TrailingSlash(t *testing.T) {
	ctx := context.Background()
	_, vclient, err := StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Should handle trailing slash without error
	err = vclient.CreateKVMountIfNotExist("secret/")
	assert.NoError(t, err)

	mounts, err := vclient.client.Sys().ListMounts()
	assert.NoError(t, err)
	assert.Contains(t, mounts, "secret/")
}

func TestCreateKVMountIfNotExist_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	_, vclient, err := StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = vclient.CreateKVMountIfNotExist("secret")
	assert.NoError(t, err)

	// Calling again should be a no-op, not an error
	err = vclient.CreateKVMountIfNotExist("secret")
	assert.NoError(t, err)
}

func TestWriteSecret(t *testing.T) {
	ctx := context.Background()
	_, vclient, err := StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = vclient.CreateKVMountIfNotExist("secret")
	assert.NoError(t, err)

	err = vclient.WriteSecret("secret", "mykey", "myvalue")
	assert.NoError(t, err)

	// Read back and verify
	secret, err := vclient.client.Logical().Read("secret/data/mykey")
	assert.NoError(t, err)
	assert.NotNil(t, secret)
	data := secret.Data["data"].(map[string]interface{})
	assert.Equal(t, "myvalue", data["value"])
}

func TestWriteSecret_TrailingSlashOnMount(t *testing.T) {
	ctx := context.Background()
	_, vclient, err := StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = vclient.CreateKVMountIfNotExist("secret")
	assert.NoError(t, err)

	// Trailing slash on mountPath should not produce a double-slash path
	err = vclient.WriteSecret("secret/", "mykey", "myvalue")
	assert.NoError(t, err)

	secret, err := vclient.client.Logical().Read("secret/data/mykey")
	assert.NoError(t, err)
	assert.NotNil(t, secret)
}
