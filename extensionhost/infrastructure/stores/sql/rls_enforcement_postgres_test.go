package sql_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"testing"

	_ "github.com/lib/pq"

	sqlstore "github.com/movebigrocks/extension-sdk/extensionhost/infrastructure/stores/sql"
	"github.com/movebigrocks/extension-sdk/extensionhost/testutil"
)

// TestForkStoreRLSEnforcement proves the extension (copied) store honors
// row-level security: connected as a non-bypassing owner role with RLS forced on
// a core table, SetTenantContext confines reads and writes to the workspace,
// empty context is fail-safe, and WithAdminContext switches to the mbr_admin
// bypass role and sees every workspace. This mirrors the core enforcement proof
// and is the gate for enabling RLS while the extension runtimes access core
// tables through this store.
func TestForkStoreRLSEnforcement(t *testing.T) {
	testDSN, cleanupDB := testutil.SetupTestPostgresDatabase(t)
	defer cleanupDB()

	suDB, err := sql.Open("postgres", testDSN)
	if err != nil {
		t.Fatalf("open superuser connection: %v", err)
	}
	defer suDB.Close()
	ctx := context.Background()

	appRole := "mbr_ext_rls_" + roleSuffix(testDSN)
	seedRoles(t, suDB, appRole, dbNameFromDSN(t, testDSN))
	defer dropRole(suDB, appRole)

	// Open the store as the application role. The migration runner builds and
	// owns the schema, so the role is subject to forced row-level security.
	store, closeStore := openStoreAs(t, testDSN, appRole)
	defer closeStore()

	// Force RLS on the core contacts table with the workspace policy.
	for _, stmt := range []string{
		`CREATE OR REPLACE FUNCTION public.current_workspace_id() RETURNS UUID LANGUAGE plpgsql STABLE SECURITY DEFINER AS $$ BEGIN RETURN NULLIF(current_setting('app.current_workspace_id', true), '')::uuid; END; $$;`,
		`ALTER TABLE core_platform.contacts ENABLE ROW LEVEL SECURITY`,
		`ALTER TABLE core_platform.contacts FORCE ROW LEVEL SECURITY`,
		`DROP POLICY IF EXISTS tenant_isolation ON core_platform.contacts`,
		`CREATE POLICY tenant_isolation ON core_platform.contacts FOR ALL USING (workspace_id = public.current_workspace_id())`,
	} {
		if _, err := suDB.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("enable rls stmt %q: %v", stmt, err)
		}
	}

	wsA := seedWorkspaceContacts(t, suDB, "alpha", "aa", 3)
	wsB := seedWorkspaceContacts(t, suDB, "bravo", "bb", 2)

	count := func(setup func(context.Context) error) int {
		t.Helper()
		var n int
		if err := store.WithTransaction(ctx, func(txCtx context.Context) error {
			if setup != nil {
				if err := setup(txCtx); err != nil {
					return err
				}
			}
			return store.SqlxDB().Get(txCtx).GetContext(txCtx, &n, `SELECT count(*) FROM core_platform.contacts`)
		}); err != nil {
			t.Fatalf("count: %v", err)
		}
		return n
	}

	if got := count(func(c context.Context) error { return store.SetTenantContext(c, wsA) }); got != 3 {
		t.Fatalf("workspace A: want 3, got %d", got)
	}
	if got := count(func(c context.Context) error { return store.SetTenantContext(c, wsB) }); got != 2 {
		t.Fatalf("workspace B: want 2, got %d", got)
	}
	if got := count(nil); got != 0 {
		t.Fatalf("no context: want 0 (fail-safe), got %d", got)
	}

	var adminCount int
	if err := store.WithAdminContext(ctx, func(txCtx context.Context) error {
		return store.SqlxDB().Get(txCtx).GetContext(txCtx, &adminCount, `SELECT count(*) FROM core_platform.contacts`)
	}); err != nil {
		t.Fatalf("admin context: %v", err)
	}
	if adminCount != 5 {
		t.Fatalf("admin context: want 5, got %d", adminCount)
	}
}

func seedRoles(t *testing.T, db *sql.DB, appRole, dbName string) {
	t.Helper()
	ctx := context.Background()
	stmts := []string{
		`DO $$ BEGIN IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname='mbr_admin') THEN CREATE ROLE mbr_admin BYPASSRLS; ELSE ALTER ROLE mbr_admin BYPASSRLS; END IF; END $$;`,
		fmt.Sprintf(`CREATE ROLE %s LOGIN PASSWORD 'rls_test_pw' NOBYPASSRLS`, appRole),
		fmt.Sprintf(`GRANT mbr_admin TO %s`, appRole),
		fmt.Sprintf(`GRANT CREATE ON DATABASE %s TO %s`, dbName, appRole),
		fmt.Sprintf(`GRANT CREATE, USAGE ON SCHEMA public TO %s`, appRole),
	}
	for _, s := range stmts {
		if _, err := db.ExecContext(ctx, s); err != nil {
			t.Fatalf("seed role %q: %v", s, err)
		}
	}
}

func dropRole(db *sql.DB, appRole string) {
	ctx := context.Background()
	_, _ = db.ExecContext(ctx, fmt.Sprintf(`REVOKE mbr_admin FROM %s`, appRole))
	_, _ = db.ExecContext(ctx, fmt.Sprintf(`DROP OWNED BY %s`, appRole))
	_, _ = db.ExecContext(ctx, fmt.Sprintf(`DROP ROLE IF EXISTS %s`, appRole))
}

func seedWorkspaceContacts(t *testing.T, db *sql.DB, slug, code string, n int) string {
	t.Helper()
	ctx := context.Background()
	var ws string
	if err := db.QueryRowContext(ctx,
		`INSERT INTO core_platform.workspaces (name, slug, short_code) VALUES ($1,$2,$3) RETURNING id`,
		slug, slug, code).Scan(&ws); err != nil {
		t.Fatalf("seed workspace: %v", err)
	}
	for i := 0; i < n; i++ {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO core_platform.contacts (workspace_id, email) VALUES ($1,$2)`,
			ws, fmt.Sprintf("%s-%d@example.com", slug, i)); err != nil {
			t.Fatalf("seed contact: %v", err)
		}
	}
	return ws
}

func openStoreAs(t *testing.T, adminDSN, role string) (*sqlstore.Store, func()) {
	t.Helper()
	parsed, err := url.Parse(adminDSN)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	parsed.User = url.UserPassword(role, "rls_test_pw")
	db, err := sqlstore.NewDBWithConfig(sqlstore.DBConfig{DSN: parsed.String()})
	if err != nil {
		t.Fatalf("open store as %s: %v", role, err)
	}
	store, err := sqlstore.NewStore(db)
	if err != nil {
		db.Close()
		t.Fatalf("new store: %v", err)
	}
	return store, func() { _ = store.Close(); _ = db.Close() }
}

func dbNameFromDSN(t *testing.T, dsn string) string {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	return strings.TrimPrefix(parsed.Path, "/")
}

func roleSuffix(dsn string) string {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return "x"
	}
	return strings.ToLower(strings.ReplaceAll(strings.TrimPrefix(parsed.Path, "/"), "-", ""))
}
