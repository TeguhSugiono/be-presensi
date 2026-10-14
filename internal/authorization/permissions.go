package authorization

import "strings"

type Role string
type Permission string

const (
	RoleAdmin    Role = "admin"
	RoleGuru     Role = "guru"
	RoleKaryawan Role = "karyawan"
)

const (
	DashboardView           Permission = "dashboard.view"
	AttendanceView          Permission = "attendance.view"
	ScheduleManage          Permission = "schedule.manage"
	DailyTaskManage         Permission = "daily_task.manage"
	QuarterTaskMasterManage Permission = "quarter_task_master.manage"
	QuarterTaskView         Permission = "quarter_task.view"
	AttendanceReportView    Permission = "attendance_report.view"
	DailyTaskReportView     Permission = "daily_task_report.view"
	InboxView               Permission = "inbox.view"
	ProfileView             Permission = "profile.view"

	UsersView   Permission = "users.view"
	UsersCreate Permission = "users.create"
	UsersUpdate Permission = "users.update"
	UsersDelete Permission = "users.delete"

	RolesView   Permission = "roles.view"
	RolesCreate Permission = "roles.create"
	RolesUpdate Permission = "roles.update"
	RolesDelete Permission = "roles.delete"
)

// NormalizeRole menormalkan nama role dari database atau login.
func NormalizeRole(role string) Role {
	return Role(strings.ToLower(strings.TrimSpace(role)))
}

// rolePermissions adalah sumber aturan hak akses backend.
var rolePermissions = map[Role]map[Permission]bool{
	RoleAdmin: {
		DashboardView:           true,
		ScheduleManage:          true,
		QuarterTaskMasterManage: true,
		AttendanceReportView:    true,
		DailyTaskReportView:     true,
		ProfileView:             true,

		UsersView:   true,
		UsersCreate: true,
		UsersUpdate: true,
		UsersDelete: true,

		RolesView:   true,
		RolesCreate: true,
		RolesUpdate: true,
		RolesDelete: true,
	},

	RoleGuru: {
		DashboardView:   true,
		AttendanceView:  true,
		DailyTaskManage: true,
		QuarterTaskView: true,
		InboxView:       true,
		ProfileView:     true,
	},

	RoleKaryawan: {
		DashboardView:   true,
		AttendanceView:  true,
		DailyTaskManage: true,
		QuarterTaskView: true,
		InboxView:       true,
		ProfileView:     true,
	},
}

// HasPermission memeriksa apakah role memiliki permission tertentu.
func HasPermission(role string, permission Permission) bool {
	normalizedRole := NormalizeRole(role)
	permissions, exists := rolePermissions[normalizedRole]
	if !exists {
		return false
	}

	return permissions[permission]
}
