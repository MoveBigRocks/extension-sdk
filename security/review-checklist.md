# Extension Review Checklist

Mark each item before production activation.

- Manifest fields are renamed from the sample defaults.
- `mbr extensions lint . --json` passes.
- `extension.contract.json` exists and matches the intended extension surface.
- Routes and asset paths are valid.
- The extension installs from the source directory.
- `mbr extensions validate --id EXTENSION_ID` passes.
- `mbr extensions show --id EXTENSION_ID --json` shows the expected resolved navigation, widgets, and seeded resources.
- `mbr extensions monitor --id EXTENSION_ID` reports healthy.
- The main admin and public workflows were exercised in a sandbox workspace.
- The threat model is complete.
- The extension does not request privileged behavior unnecessarily.
- Rollback and deactivate steps are known.
- The instance repo records how this extension should be installed and configured.
