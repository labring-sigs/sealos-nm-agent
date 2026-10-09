package pod

import (
	"context"
	"testing"
)

func TestGetPodTypeAndTypeName(t *testing.T) {
	tests := []struct {
		name       string
		labels     map[string]string
		wantType   POD_TYPE
		wantTypeName string
	}{
		{
			name:     "nil labels fall back to other",
			labels:   nil,
			wantType: POD_TYPE_OTHER,
		},
		{
			name: "kubeblocks database",
			labels: map[string]string{
				"apps.kubeblocks.io/component-name": "mysql",
				"app.kubernetes.io/instance":        "test-db",
			},
			wantType:     POD_TYPE_DB,
			wantTypeName: "test-db",
		},
		{
			name: "terminal",
			labels: map[string]string{
				"TerminalID": "some-terminal-id",
			},
			wantType: POD_TYPE_TERMINAL,
		},
		{
			// regression: devbox pods used to be classified as OTHER with an
			// empty name, which never matched the billing monitor's DEV-BOX
			// combo, so their traffic was silently unbilled
			name: "devbox",
			labels: map[string]string{
				"app.kubernetes.io/managed-by": "sealos",
				"app.kubernetes.io/name":       "jk10",
				"app.kubernetes.io/part-of":    "devbox",
			},
			wantType:     POD_TYPE_DEVBOX,
			wantTypeName: "jk10",
		},
		{
			name: "devbox without name label",
			labels: map[string]string{
				"app.kubernetes.io/part-of": "devbox",
			},
			wantType: POD_TYPE_DEVBOX,
		},
		{
			name: "part-of label with other value is not devbox",
			labels: map[string]string{
				"app.kubernetes.io/part-of": "something-else",
			},
			wantType: POD_TYPE_OTHER,
		},
		{
			name: "app",
			labels: map[string]string{
				"app": "main-realtime-v3",
			},
			wantType:     POD_TYPE_APP,
			wantTypeName: "main-realtime-v3",
		},
		{
			name: "job",
			labels: map[string]string{
				"job-name": "myjob-abc123",
			},
			wantType:     POD_TYPE_JOB,
			wantTypeName: "myjob",
		},
		{
			name:     "unlabeled pod falls back to other",
			labels:   map[string]string{"foo": "bar"},
			wantType: POD_TYPE_OTHER,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotTypeName := GetPodTypeAndTypeName(ctx, tt.labels)
			if gotType != tt.wantType {
				t.Errorf("GetPodTypeAndTypeName() type = %v, want %v", gotType, tt.wantType)
			}
			if gotTypeName != tt.wantTypeName {
				t.Errorf("GetPodTypeAndTypeName() typeName = %q, want %q", gotTypeName, tt.wantTypeName)
			}
		})
	}
}

// devbox must carry the same numeric value as the billing monitor's devBox
// enum; the stored pod_type is exact-matched when traffic is attributed.
func TestPodTypeValuesStayInSyncWithBilling(t *testing.T) {
	if POD_TYPE_DEVBOX != 10 {
		t.Errorf("POD_TYPE_DEVBOX = %v, want 10 (billing app type devBox)", POD_TYPE_DEVBOX)
	}
	if POD_TYPE_OTHER != 5 {
		t.Errorf("POD_TYPE_OTHER = %v, want 5 (billing app type other)", POD_TYPE_OTHER)
	}
}
