package enum

type UserRole string

const (
	RoleAdmin            UserRole = "admin"
	RoleEventCoordinator UserRole = "event_coordinator"
	RoleUser             UserRole = "user"
)

func (r UserRole) String() string {
	return string(r)
}
