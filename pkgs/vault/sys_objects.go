package vault

type AuthRole struct {
	BoundServiceAccounts []string `json:"bound_service_accounts"`
	MaxJWTExp            int      `json:"max_jwt_exp"`
	Policies             []string `json:"policies"`
}

type AuthRoleResp struct {
	Data AuthRole `json:"data"`
}

type GeneralResp struct {
	Data map[string]interface{} `json:"data"`
}
