# Tasks

## TODO

- [x] **Audit existing docs**: Review profile-related docs in:
  - [x] `glazed/pkg/doc/topics/12-profiles-use-code.md`
  - [x] `glazed/pkg/doc/topics/21-cmds-middlewares.md` (profile snippets)
  - [x] `geppetto/pkg/doc/topics/01-profiles.md`
- [x] **Add new Glazed doc page**: Create `glazed/pkg/doc/topics/15-profiles.md` explaining:
  - [x] `profiles.yaml` format (profile → layer slug → param → value)
  - [x] default location and XDG: `$XDG_CONFIG_HOME/<app>/profiles.yaml` (fallback: `~/.config/...` via `os.UserConfigDir`)
  - [x] selection mechanisms: `--profile`, `--profile-file`, `APP_PROFILE`, `APP_PROFILE_FILE`, config file keys under `profile-settings`
  - [x] precedence: flags > env > config > profiles > defaults
  - [x] required behavior: unknown profile should error; missing default profile file errors only if non-default profile is requested
  - [x] how to enable flags: `cli.WithProfileSettingsLayer()` when building Cobra commands
  - [x] how to debug: `--print-parsed-parameters`
- [x] **Update existing Glazed docs**:
  - [x] Fix `12-profiles-use-code.md` to stop being Pinocchio-specific, reference the new doc, and describe bootstrap selection correctly.
  - [x] Fix/clarify profile examples in `21-cmds-middlewares.md` (correct signature + ordering + selection caveat).
- [x] **Update Geppetto docs**:
  - [x] Fix `geppetto/pkg/doc/topics/01-profiles.md` paths (`~/.pinocchio/config.yaml` vs `~/.config/pinocchio/config.yaml` etc)
  - [x] Fix config example to use `profile-settings:` layer keys (not top-level `profile:`)
  - [x] Mention `PINOCCHIO_PROFILE_FILE` and `--profile-file`
  - [x] Mention precedence and `--print-parsed-parameters`
- [x] **Relate files + close ticket**: Add RelatedFiles notes and close ticket with a changelog entry.

