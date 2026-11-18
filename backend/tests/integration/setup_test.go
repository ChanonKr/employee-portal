//go:build integration

// Package integration runs the real repositories, migrations and HTTP server
// against PostgreSQL 17 in a container. Run with: make test-integration
package integration

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"

	profilerepo "employee-portal/modules/profile/repo"
	profileservice "employee-portal/modules/profile/service"
	"employee-portal/pkg/database"
	"employee-portal/pkg/imaging"
	"employee-portal/pkg/storage"
)

var (
	ownerPool   *pgxpool.Pool // schema owner: raw SQL to arrange and inspect test data
	appPool     *pgxpool.Pool // role "app", raw SQL: checks what the role may not do
	appDB       *gorm.DB      // role "app" through GORM: what the API uses in production
	retentionDB *gorm.DB      // role "retention": what the retention job uses
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()

	ctr, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("portal"),
		tcpostgres.WithUsername("owner"),
		tcpostgres.WithPassword("owner"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		log.Printf("start postgres: %v", err)
		return 1
	}
	defer func() { _ = testcontainers.TerminateContainer(ctr) }()

	ownerURL, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("connection string: %v", err)
		return 1
	}

	mig, err := migrate.New("file://../../pkg/database/migrations", "pgx5"+ownerURL[len("postgres"):])
	if err != nil {
		log.Printf("migrate init: %v", err)
		return 1
	}
	if err := mig.Up(); err != nil {
		log.Printf("migrate up: %v", err)
		return 1
	}

	if ownerPool, err = pgxpool.New(ctx, ownerURL); err != nil {
		log.Printf("owner pool: %v", err)
		return 1
	}
	defer ownerPool.Close()

	// Migrations create "app" as NOLOGIN; give it a password just for this container.
	if _, err := ownerPool.Exec(ctx, `ALTER ROLE app LOGIN PASSWORD 'app'`); err != nil {
		log.Printf("enable app login: %v", err)
		return 1
	}
	host, _ := ctr.Host(ctx)
	port, _ := ctr.MappedPort(ctx, "5432/tcp")
	appURL := fmt.Sprintf("postgres://app:app@%s:%s/portal?sslmode=disable", host, port.Port())
	if appPool, err = pgxpool.New(ctx, appURL); err != nil {
		log.Printf("app pool: %v", err)
		return 1
	}
	defer appPool.Close()
	if appDB, err = database.Open(ctx, appURL); err != nil {
		log.Printf("app gorm: %v", err)
		return 1
	}
	defer database.Close(appDB)

	if _, err := ownerPool.Exec(ctx, `ALTER ROLE retention LOGIN PASSWORD 'retention'`); err != nil {
		log.Printf("enable retention login: %v", err)
		return 1
	}
	retentionURL := fmt.Sprintf("postgres://retention:retention@%s:%s/portal?sslmode=disable", host, port.Port())
	if retentionDB, err = database.Open(ctx, retentionURL); err != nil {
		log.Printf("retention gorm: %v", err)
		return 1
	}
	defer database.Close(retentionDB)

	return m.Run()
}

// newEmployee inserts a fresh user + employee so tests do not share state.
func newEmployee(t *testing.T, mobile string) (userID, employeeID int64, code string) {
	t.Helper()
	ctx := context.Background()
	code = fmt.Sprintf("T-%d", time.Now().UnixNano()%1_000_000_000_000)
	err := ownerPool.QueryRow(ctx, `
		WITH u AS (
		  INSERT INTO users (employee_code, password_hash)
		  VALUES ($1, '$2a$10$NQ.nWEZwH8//fqJvQwKneub8Nkn7s1pH91J2sPwMMz3ProcdR9y1S') -- demo1234
		  RETURNING id)
		INSERT INTO employees (user_id, first_name, last_name, mobile_phone)
		SELECT id, 'Test', 'User', $2 FROM u
		RETURNING user_id, id`, code, mobile).Scan(&userID, &employeeID)
	if err != nil {
		t.Fatalf("insert employee: %v", err)
	}
	return userID, employeeID, code
}

func countLogs(t *testing.T, employeeID int64) int {
	t.Helper()
	var n int
	if err := ownerPool.QueryRow(context.Background(),
		`SELECT count(*) FROM profile_change_logs WHERE employee_id = $1`, employeeID).Scan(&n); err != nil {
		t.Fatalf("count logs: %v", err)
	}
	return n
}

// newProfileService wires the real repository, image processor and local storage
// (in a temp dir) — the same pieces the server uses.
func newProfileService(t *testing.T) *profileservice.Service {
	uc, _ := newProfileServiceWithDir(t)
	return uc
}

func newProfileServiceWithDir(t *testing.T) (*profileservice.Service, string) {
	t.Helper()
	dir := t.TempDir()
	files, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatalf("storage: %v", err)
	}
	return profileservice.New(profileservice.Deps{
		Repo:   profilerepo.New(appDB),
		Photos: imaging.NewProcessor(5 << 20),
		Files:  files,
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}), dir
}
