#!/usr/bin/env bash
set -euo pipefail

# check-core-security-parity.sh
#
# The extensionhost tree in this SDK is a copy of platform/pkg/extensionhost.
# While that copy exists, security-critical guards must not silently drift from
# the platform source of truth. This is exactly how the extension schema
# migration containment guard was lost once: the copy predated the guard, so
# the vendored path could run DROP, ALTER, GRANT, or writes against core_
# schemas. This gate fails the build if a listed guard is missing from the copy
# or differs from the platform original.
#
# It intentionally checks specific guards rather than whole files, because the
# copy is being dismantled and is not expected to match the platform tree in
# full. New security-critical regions should be added to GUARDS below.

SDK_ROOT="${SDK_ROOT:-$(git rev-parse --show-toplevel)}"

# Platform is checked out at ./platform in CI (see .github/workflows/validate.yml)
# and lives beside this repo in a local workspace checkout.
if [ -n "${PLATFORM_ROOT:-}" ]; then
	:
elif [ -d "${SDK_ROOT}/platform/pkg/extensionhost" ]; then
	PLATFORM_ROOT="${SDK_ROOT}/platform"
elif [ -d "$(dirname "${SDK_ROOT}")/platform/pkg/extensionhost" ]; then
	PLATFORM_ROOT="$(dirname "${SDK_ROOT}")/platform"
else
	echo "FAIL: cannot locate the platform checkout; set PLATFORM_ROOT" >&2
	exit 1
fi

fail=0
tmp="$(mktemp -d)"
trap 'rm -rf "${tmp}"' EXIT

# extract_region <awk-start-regex> <awk-end-regex> <file>
# Prints the contiguous block from the first line matching start through the
# first line at column zero matching end (inclusive).
extract_region() {
	awk -v start="$1" -v end="$2" '
		$0 ~ start { p = 1 }
		p { print }
		p && $0 ~ end && seen_body { exit }
		p { seen_body = 1 }
	' "$3"
}

# GUARDS: one entry per security-critical region.
#   label | relative path under the extensionhost tree | start regex | end regex
GUARDS=(
	"extension-migration-containment|infrastructure/stores/sql/extension_schema_migrator.go|^// protectedSchemaTarget matches|^func isIdentRune"
	# The bundle signer (scripts/sign-bundle) signs the exact bytes VerifyBundle
	# checks, by reusing this package's CanonicalSignedBundlePayload and
	# BundleLicenseClaim. If the license-claim shape or the canonicalization drift
	# from the platform verifier, a bundle the SDK signs is one the core rejects.
	# This is the drift that broke every signature once: the signer's own copy of
	# the license claim marked every field omitempty, so a public bundle's license
	# serialized shorter than the verifier reconstructed it.
	"bundle-license-claim-shape|platform/services/extension_bundle_verifier.go|^type bundleLicenseClaim struct|^}"
	"bundle-signed-payload-canonicalization|platform/services/extension_bundle_verifier.go|^func canonicalSignedBundlePayload|^func checksumSHA256Hex"
)

for entry in "${GUARDS[@]}"; do
	IFS='|' read -r label rel start end <<<"${entry}"
	sdk_file="${SDK_ROOT}/extensionhost/${rel}"
	plat_file="${PLATFORM_ROOT}/pkg/extensionhost/${rel}"

	for f in "${sdk_file}" "${plat_file}"; do
		if [ ! -f "${f}" ]; then
			echo "FAIL[${label}]: missing file ${f}"
			fail=1
			continue 2
		fi
	done

	extract_region "${start}" "${end}" "${plat_file}" >"${tmp}/plat" || true
	extract_region "${start}" "${end}" "${sdk_file}" >"${tmp}/sdk" || true

	if [ ! -s "${tmp}/plat" ]; then
		echo "FAIL[${label}]: guard region not found in platform source ${plat_file}"
		fail=1
		continue
	fi
	if [ ! -s "${tmp}/sdk" ]; then
		echo "FAIL[${label}]: guard region missing from SDK copy ${sdk_file}"
		fail=1
		continue
	fi
	if ! diff -u "${tmp}/plat" "${tmp}/sdk" >"${tmp}/diff"; then
		echo "FAIL[${label}]: SDK guard drifted from platform source:"
		cat "${tmp}/diff"
		fail=1
		continue
	fi
	echo "OK[${label}]: guard is present and identical to the platform core"
done

# The migrator must actually invoke the containment guard before executing SQL.
mig="${SDK_ROOT}/extensionhost/infrastructure/stores/sql/extension_schema_migrator.go"
if ! grep -q 'validateExtensionMigrationContainment(migration.SQL)' "${mig}"; then
	echo "FAIL: extension_schema_migrator.go does not call validateExtensionMigrationContainment"
	fail=1
else
	echo "OK: extension_schema_migrator.go invokes the containment guard"
fi

if [ "${fail}" -ne 0 ]; then
	echo "core security parity check failed" >&2
	exit 1
fi
echo "core security parity check passed"
