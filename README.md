# mf-digisac

[mf](https://github.com/mguilarducci/mf-cli) plugin for the Digisac API. Runs as `mf digisac`, or standalone as `mf-digisac`.

```sh
mise run install          # build and `mf plugin install ./bin/mf-digisac`
mf digisac auth login
mf digisac contacts list
```

Credentials come from `mf digisac auth login` or the `DIGISAC_URL` / `DIGISAC_TOKEN` env vars.
