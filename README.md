# WizQuickLauncher

> This product is now currently being updated, and under extreme development, basic login functionality works, modding capability being added soon!

## Security and encrypted logins

The configuration generator now stores account passwords encrypted. To decrypt them at runtime, set the environment variable `WIZQL_PASSPHRASE` to the passphrase you used while creating the config. The `config.json` file is written with restrictive permissions (0600) to keep secrets private on disk.