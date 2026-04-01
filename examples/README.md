# Local Setup and Running the example

1. From the root of the checkout, start vault server
```
docker compose -f examples/compose.yaml up -d
```

2. From the root of the checkout, run vault-init
```
go run . --initout "examples/vault_out.json" --post-init-file examples/postinit.yaml --vault-addr http://localhost:8200
```

3. If you want to test it out with gcp secrets
    - Create a gcp secret which contains the vault root token and add a version with `{"root_token": "myroot"}`
    - Create a gcp secret which will get synced with vault secret and add a version
    - In `postinit.yaml`, update `type: secret_sync` with the gcp secret path which will get synced with vault. `postinit.yaml` contains the example
    - Run vault-init with the gcp secret path like
    ```
    go run . --initout "projects/593849460306/secrets/delete_me" --post-init-file examples/postinit.yaml --vault-addr http://localhost:8200
    ```
