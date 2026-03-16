package authorization

import (
	"github.com/casbin/casbin/v2"
	"github.com/thienel/tlog"
	"go.uber.org/zap"
)

// defaultPolicies khai báo các policy mặc định
// Format: role, path, method
var defaultPolicies = [][]string{
	// SYSTEM_ADMIN = Super user (toàn quyền)
	{"SYSTEM_ADMIN", "*", "*"},

	// ADMIN permissions - User management
	{"ADMIN", "/api/users", "GET"},
	{"ADMIN", "/api/users/:id", "GET"},
	{"ADMIN", "/api/users", "POST"},
	{"ADMIN", "/api/users/:id", "PUT"},
	{"ADMIN", "/api/users/:id", "DELETE"},
}

// defaultRoleLinks khai báo role hierarchy
// Format: child, parent (child kế thừa quyền của parent)
var defaultRoleLinks = [][]string{
	// SYSTEM_ADMIN kế thừa tất cả quyền ADMIN
	{"SYSTEM_ADMIN", "ADMIN"},
}

// seedDefaultPolicies thêm policies mặc định nếu DB trống
func seedDefaultPolicies(enforcer *casbin.Enforcer) error {
	existingPolicies, err := enforcer.GetPolicy()
	if err != nil {
		return err
	}
	if len(existingPolicies) > 0 {
		tlog.Info("Casbin policies already exist, skipping seed",
			zap.Int("count", len(existingPolicies)),
		)
		return nil
	}

	tlog.Info("Seeding default Casbin policies...")

	// Add policies
	for _, p := range defaultPolicies {
		if _, err := enforcer.AddPolicy(p[0], p[1], p[2]); err != nil {
			return err
		}
	}

	// Add role hierarchy
	for _, g := range defaultRoleLinks {
		if _, err := enforcer.AddGroupingPolicy(g[0], g[1]); err != nil {
			return err
		}
	}

	tlog.Info("Default Casbin policies seeded",
		zap.Int("policies", len(defaultPolicies)),
		zap.Int("role_links", len(defaultRoleLinks)),
	)

	return nil
}
