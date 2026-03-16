package authorization

import (
	"database/sql"
	"fmt"

	"github.com/casbin/casbin/v2"
	casbinpgadapter "github.com/cychiuae/casbin-pg-adapter"
	"github.com/thienel/tlog"
	"go.uber.org/zap"
)

const casbinTableName = "casbin_rules"

// NewEnforcer creates a Casbin enforcer with PostgreSQL adapter
func NewEnforcer(modelPath string, db *sql.DB) (*casbin.Enforcer, error) {
	// Create PostgreSQL adapter
	adapter, err := casbinpgadapter.NewAdapter(db, casbinTableName)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin adapter: %w", err)
	}

	// Create enforcer from model file + adapter
	enforcer, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	// Enable auto-save: policies are persisted to DB automatically
	enforcer.EnableAutoSave(true)

	// Load policies from DB
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load casbin policies: %w", err)
	}

	// Seed default policies if DB is empty
	if err := seedDefaultPolicies(enforcer); err != nil {
		return nil, fmt.Errorf("failed to seed default policies: %w", err)
	}

	tlog.Info("Casbin enforcer initialized",
		zap.String("model", modelPath),
		zap.String("table", casbinTableName),
	)

	return enforcer, nil
}
