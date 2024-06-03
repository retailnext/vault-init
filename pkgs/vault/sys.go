package vault

import (
	"encoding/json"
	"fmt"
	"log/slog"

	vaultapi "github.com/hashicorp/vault/api"
)

func (v *Vault) GetInitStatus() (init bool, err error) {
	return v.client.Sys().InitStatusWithContext(v.ctx)
}

func (v *Vault) GetHealthStatus() (*vaultapi.HealthResponse, error) {
	return v.client.Sys().HealthWithContext(v.ctx)
}

func (v *Vault) Initialize() (resp []byte, err error) {
	initialized, err := v.client.Sys().InitStatus()
	if err != nil || initialized {
		// it is already initialized or fails to get the status
		return
	}

	initRequest := &vaultapi.InitRequest{
		StoredShares:      3,
		RecoveryShares:    3,
		RecoveryThreshold: 2,
	}
	initResponse, err := v.client.Sys().Init(initRequest)
	if err != nil {
		slog.Error("fail_to_init", "err", err)
		return
	}

	vaultRootToken := initResponse.RootToken
	if vaultRootToken == "" {
		slog.Error("no_root_token")
		return resp, fmt.Errorf("no root token is found after initialized")
	}
	v.SetToken(vaultRootToken)

	return json.Marshal(initResponse)
}

func (v *Vault) SetPolicy(policyName string, policyContent string) (err error) {
	payload := map[string]interface{}{
		"policy": policyContent,
	}
	policyPath := fmt.Sprintf("sys/policy/%s", policyName)
	resp, err := v.client.Logical().WriteWithContext(v.ctx, policyPath, payload)
	slog.Info("set_policy", "resp", fmt.Sprintf("%v", resp))
	return err
}

func (v *Vault) GetPolicy(policyName string) (policyContent string, err error) {
	policyPath := fmt.Sprintf("sys/policy/%s", policyName)
	resp, err := v.client.Logical().ReadWithContext(v.ctx, policyPath)
	slog.Debug(fmt.Sprintf("%v", resp))

	if resp.Data == nil {
		return "", fmt.Errorf("there is no data in the response from vault")
	}
	var ok bool
	if policyContent, ok = resp.Data["rules"].(string); !ok {
		return "", fmt.Errorf("policy data content is not found in the response")
	}
	return
}

func (v *Vault) EnableAuth(authMethod string) (err error) {
	// setting authMethod as auth path for convience
	return v.EnableAuthByPath(authMethod, authMethod)
}

func (v *Vault) EnableAuthByPath(authPath string, authMethod string) (err error) {
	payload := &vaultapi.EnableAuthOptions{
		Type: authMethod,
	}
	return v.client.Sys().EnableAuthWithOptionsWithContext(v.ctx, authPath, payload)
}

func (v *Vault) GetAuthType(authPath string) (authType string, err error) {
	resp, err := v.client.Logical().ReadWithContext(v.ctx, fmt.Sprintf("sys/auth/%s", authPath))
	if err != nil {
		return
	}
	if resp.Data == nil {
		return "", fmt.Errorf("there is no data in the response from vault")
	}
	var ok bool
	if authType, ok = resp.Data["type"].(string); !ok {
		return "", fmt.Errorf("policy data content is not found in the response")
	}
	return
}

func (v *Vault) AddAuthJWTConfig(authPath, oidcDiscoveryUrl, boundIssuer string) (err error) {
	payload := map[string]interface{}{
		"oidc_discovery_url": oidcDiscoveryUrl,
		"bound_issuer":       boundIssuer,
	}
	_, err = v.client.Logical().WriteWithContext(v.ctx, fmt.Sprintf("auth/%s/config", authPath), payload)
	return err
}

func (v *Vault) AddAuthRoleIAMType(authPath string, roleName string, policies []string, serviceAccounts []string) (err error) {
	// Assume authMethod as auth path
	payload := map[string]interface{}{
		"type":                   "iam",
		"policies":               policies,
		"bound_service_accounts": serviceAccounts,
		"max_jwt_exp":            3600,
	}
	rolePath := fmt.Sprintf("auth/%s/role/%s", authPath, roleName)
	_, err = v.client.Logical().WriteWithContext(v.ctx, rolePath, payload)
	return err
}

func (v *Vault) GetAuthRole(authPath string, roleName string) (authRole AuthRole, err error) {
	rolePath := fmt.Sprintf("auth/%s/role/%s", authPath, roleName)
	resp, err := v.client.Logical().ReadRawWithContext(v.ctx, rolePath)
	if err != nil {
		return authRole, err
	}

	defer resp.Body.Close()

	authRoleResp := &AuthRoleResp{}
	err = json.NewDecoder(resp.Body).Decode(authRoleResp)
	if err != nil {
		return authRole, err
	}

	return authRoleResp.Data, nil
}

func (v *Vault) AddAuthRoleJWTType(authPath, roleName string, policies []string, boundAudiences []string, sub string, userClaim string, ttl string) (err error) {
	// Assume authMethod as auth path
	payload := map[string]interface{}{
		"role_type":         "jwt",
		"policies":          policies,
		"bound_audiences":   boundAudiences,
		"bound_claims_type": "glob",
		"bound_claims": map[string]interface{}{
			"sub": sub,
		},
		"user_claim": userClaim,
		"token_ttl":  ttl,
	}
	rolePath := fmt.Sprintf("auth/%s/role/%s", authPath, roleName)
	_, err = v.client.Logical().WriteWithContext(v.ctx, rolePath, payload)
	return err
}

func (v *Vault) GetRawAuthRole(authMethod, roleName string) (authRole map[string]interface{}, err error) {
	rolePath := fmt.Sprintf("auth/%s/role/%s", authMethod, roleName)
	resp, err := v.client.Logical().ReadRawWithContext(v.ctx, rolePath)
	if err != nil {
		return authRole, err
	}

	defer resp.Body.Close()

	msgData := &GeneralResp{}
	err = json.NewDecoder(resp.Body).Decode(msgData)
	if err != nil {
		return
	}
	authRole = msgData.Data
	return
}
