# Local Setup and Running the example

1. From the root of the checkout, start vault server
```
docker compose -f examples/compose.yaml up -d
```

2. From the root of the checkout, run vault-init
```
go run . --initout "examples/vault_out.json" --post-init-file examples/postinit.yaml --vault-addr http://localhost:8200
```
