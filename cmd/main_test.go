package main

import (
	"testing"
)

func TestAnalyzeContainers(t *testing.T) {
	tests := []struct {
		name        string
		containers  []ContainerInfo
		expectCount int
	}{
		{
			name:        "empty container list",
			containers:  []ContainerInfo{},
			expectCount: 0,
		},
		{
			name: "healthy running container",
			containers: []ContainerInfo{
				{
					Name: "nginx",
					State: struct {
						Status       string
						Running      bool
						RestartCount int    `json:"RestartCount"`
						StartedAt    string `json:"StartedAt"`
					}{
						Status:       "running",
						Running:      true,
						RestartCount: 0,
					},
					Config: struct{ Image string }{
						Image: "nginx:1.21",
					},
				},
			},
			expectCount: 1,
		},
		{
			name: "container with high restart count",
			containers: []ContainerInfo{
				{
					Name: "unstable-app",
					State: struct {
						Status       string
						Running      bool
						RestartCount int    `json:"RestartCount"`
						StartedAt    string `json:"StartedAt"`
					}{
						Status:       "running",
						Running:      true,
						RestartCount: 10,
					},
					Config: struct{ Image string }{
						Image: "app:latest",
					},
				},
			},
			expectCount: 1,
		},
		{
			name: "stopped container with restart policy always",
			containers: []ContainerInfo{
				{
					Name: "crashed-app",
					State: struct {
						Status       string
						Running      bool
						RestartCount int    `json:"RestartCount"`
						StartedAt    string `json:"StartedAt"`
					}{
						Status:       "exited",
						Running:      false,
						RestartCount: 5,
					},
					Config: struct{ Image string }{
						Image: "app:v1",
					},
					HostConfig: struct {
						RestartPolicy struct{ Name string }
					}{
						RestartPolicy: struct{ Name string }{
							Name: "always",
						},
					},
				},
			},
			expectCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reports := analyzeContainers(tt.containers)
			if len(reports) != tt.expectCount {
				t.Errorf("expected %d reports, got %d", tt.expectCount, len(reports))
			}
		})
	}
}

func TestContainerHealthStatus(t *testing.T) {
	container := ContainerInfo{
		Name: "test-container",
		State: struct {
			Status       string
			Running      bool
			RestartCount int    `json:"RestartCount"`
			StartedAt    string `json:"StartedAt"`
		}{
			Status:       "running",
			Running:      true,
			RestartCount: 0,
		},
		Config: struct{ Image string }{
			Image: "test:v1",
		},
	}

	reports := analyzeContainers([]ContainerInfo{container})
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}

	if reports[0].Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", reports[0].Status)
	}

	if len(reports[0].Issues) != 0 {
		t.Errorf("expected 0 issues for healthy container, got %d", len(reports[0].Issues))
	}
}

func TestContainerWithLatestTag(t *testing.T) {
	container := ContainerInfo{
		Name: "test-container",
		State: struct {
			Status       string
			Running      bool
			RestartCount int    `json:"RestartCount"`
			StartedAt    string `json:"StartedAt"`
		}{
			Status:       "running",
			Running:      true,
			RestartCount: 0,
		},
		Config: struct{ Image string }{
			Image: "test:latest",
		},
	}

	reports := analyzeContainers([]ContainerInfo{container})
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}

	if len(reports[0].Issues) != 1 {
		t.Errorf("expected 1 issue for latest tag, got %d", len(reports[0].Issues))
	}
}

func TestContainerNameTrimSlash(t *testing.T) {
	container := ContainerInfo{
		Name: "/my-container",
		State: struct {
			Status       string
			Running      bool
			RestartCount int    `json:"RestartCount"`
			StartedAt    string `json:"StartedAt"`
		}{
			Status:       "running",
			Running:      true,
			RestartCount: 0,
		},
		Config: struct{ Image string }{
			Image: "test:v1",
		},
	}

	reports := analyzeContainers([]ContainerInfo{container})
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}

	if reports[0].Container.Name != "my-container" {
		t.Errorf("expected name 'my-container' (without slash), got '%s'", reports[0].Container.Name)
	}
}

func TestHighRestartCountWarning(t *testing.T) {
	container := ContainerInfo{
		Name: "unstable",
		State: struct {
			Status       string
			Running      bool
			RestartCount int    `json:"RestartCount"`
			StartedAt    string `json:"StartedAt"`
		}{
			Status:       "running",
			Running:      true,
			RestartCount: 6,
		},
		Config: struct{ Image string }{
			Image: "test:v1",
		},
	}

	reports := analyzeContainers([]ContainerInfo{container})
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}

	// High restart count (6) should produce warning
	if reports[0].Status != "warning" {
		t.Errorf("expected status 'warning' for restart count > 5, got '%s'", reports[0].Status)
	}
}

func TestStoppedContainerWithRestartPolicy(t *testing.T) {
	container := ContainerInfo{
		Name: "crashed",
		State: struct {
			Status       string
			Running      bool
			RestartCount int    `json:"RestartCount"`
			StartedAt    string `json:"StartedAt"`
		}{
			Status:       "exited",
			Running:      false,
			RestartCount: 10,
		},
		Config: struct{ Image string }{
			Image: "test:v1",
		},
		HostConfig: struct {
			RestartPolicy struct{ Name string }
		}{
			RestartPolicy: struct{ Name string }{
				Name: "always",
			},
		},
	}

	reports := analyzeContainers([]ContainerInfo{container})
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}

	if reports[0].Status != "critical" {
		t.Errorf("expected status 'critical' for stopped container with restart policy, got '%s'", reports[0].Status)
	}

	if len(reports[0].Suggestions) == 0 {
		t.Error("expected suggestions for critical container")
	}
}
