package pod

import (
	"context"
	"strings"
)

type POD_TYPE int64

const (
	POD_TYPE_UNKNOWN POD_TYPE = iota
	POD_TYPE_DB
	POD_TYPE_APP
	POD_TYPE_TERMINAL
	POD_TYPE_JOB
	POD_TYPE_OTHER
	POD_TYPE_OBJECTSTORAGE
	// The numeric values must stay in sync with the app type enum used by the
	// billing monitor (labring/sealos controllers/pkg/resources/resources.go):
	// the value written to traffic_meta.pod_type is matched exactly when usage
	// is billed, so a pod classified differently here is silently unbilled.
	POD_TYPE_DEVBOX POD_TYPE = 10

	// Every CHECK_*_LABEL_KEY / *_TYPE_LABEL_KEY read by GetPodTypeAndTypeName
	// must also be listed in cachedPodLabelKeys
	// (internal/k8s_watcher/cache.go); the informer cache transform strips all
	// other labels before Reconcile sees them, which silently forces the
	// affected pod type to fall back to OTHER.
	CHECK_DB_LABEL_KEY       = "apps.kubeblocks.io/component-name"
	CHECK_TERMINAL_LABEL_KEY = "TerminalID"
	CHECK_APP_LABEL_KEY      = "app"
	CHECK_JOB_LABEL_KEY      = "job-name"
	CHECK_DEVBOX_LABEL_KEY   = "app.kubernetes.io/part-of"
	DEVBOX_LABEL_VALUE       = "devbox"
	DB_TYPE_LABEL_KEY        = "app.kubernetes.io/instance"
	APP_TYPE_LABEL_KEY       = "app"
	JOB_TYPE_LABEL_KEY       = "job-name"
	DEVBOX_TYPE_LABEL_KEY    = "app.kubernetes.io/name"
)

// GetPodTypeAndTypeName maps pod labels to the billing app type. Any label key
// consulted here must be kept in cachedPodLabelKeys, otherwise the label never
// reaches this function and the pod is billed as OTHER.
func GetPodTypeAndTypeName(ctx context.Context, labels map[string]string) (POD_TYPE, string) {
	var podTypeName string
	if dbID, isDB := labels[CHECK_DB_LABEL_KEY]; isDB && dbID != "" {
		if name, exists := labels[DB_TYPE_LABEL_KEY]; exists {
			podTypeName = name
		}
		return POD_TYPE_DB, podTypeName
	} else if tid, isTerm := labels[CHECK_TERMINAL_LABEL_KEY]; isTerm && tid != "" {
		return POD_TYPE_TERMINAL, podTypeName
	} else if partOf, hasPartOf := labels[CHECK_DEVBOX_LABEL_KEY]; hasPartOf && partOf == DEVBOX_LABEL_VALUE {
		return POD_TYPE_DEVBOX, labels[DEVBOX_TYPE_LABEL_KEY]
	} else if aid, isApp := labels[CHECK_APP_LABEL_KEY]; isApp && aid != "" {
		podTypeName = aid
		return POD_TYPE_APP, podTypeName
	} else if jid, isJob := labels[CHECK_JOB_LABEL_KEY]; isJob && jid != "" {
		podTypeName = strings.SplitN(jid, "-", 2)[0]
		return POD_TYPE_JOB, podTypeName
	}
	return POD_TYPE_OTHER, podTypeName
}
