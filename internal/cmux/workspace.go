package cmux

// CreateWorkspace asks cmux to spawn a new workspace rooted at cwd.
func CreateWorkspace(cwd string) (string, error) {
	var resp struct {
		WorkspaceID string `json:"workspace_id"`
	}
	if err := Call("workspace.create", map[string]string{"cwd": cwd}, &resp); err != nil {
		return "", err
	}
	return resp.WorkspaceID, nil
}

// CapturePane snapshots the current text on a cmux surface (terminal pane).
func CapturePane(surfaceID string) (string, error) {
	var resp struct {
		Text string `json:"text"`
	}
	if err := Call("surface.capture_pane", map[string]string{"surface_id": surfaceID}, &resp); err != nil {
		return "", err
	}
	return resp.Text, nil
}
