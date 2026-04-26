package cmux

func SidebarSetStatus(workspaceID, text, icon string) error {
	return Call("sidebar.set_status", map[string]string{
		"workspace_id": workspaceID,
		"text":         text,
		"icon":         icon,
	}, nil)
}

// SidebarLog level: "info" | "warning" | "error"
func SidebarLog(level, message string) error {
	return Call("sidebar.log", map[string]string{
		"level":   level,
		"message": message,
	}, nil)
}
