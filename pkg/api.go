package pkg

// The types in this file describe the subset of the PD and TiCDC HTTP API
// payloads that ticheck reads. They used to be copied verbatim from
// pingcap/pd and pingcap/kvproto, which pulled etcd, gRPC and the whole PD
// server into go.mod for the sake of a dozen fields. Only what the checks
// actually consume is kept here; every other field in the response is ignored
// by encoding/json, so a trimmed struct decodes exactly the same values.

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Duration mirrors pd's typeutil.Duration: it is carried as a Go duration
// string such as "30s" in the JSON payload.
type Duration struct {
	time.Duration
}

// MarshalJSON writes the duration as a quoted string.
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(d.String())), nil
}

// UnmarshalJSON parses a quoted duration string (e.g. "1s", "30m").
func (d *Duration) UnmarshalJSON(text []byte) error {
	s, err := strconv.Unquote(string(text))
	if err != nil {
		return fmt.Errorf("duration must be a quoted string, got %s: %w", text, err)
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("parse duration %q failed: %w", s, err)
	}
	d.Duration = dur
	return nil
}

// StringSlice mirrors pd's typeutil.StringSlice: PD sends a list such as
// location-labels as one comma separated string ("zone,host").
type StringSlice []string

// MarshalJSON writes the slice as a quoted, comma separated string.
func (s StringSlice) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(strings.Join(s, ","))), nil
}

// UnmarshalJSON parses either a quoted, comma separated string (what PD
// sends today) or a plain JSON array, so a future PD that switches to the
// array form keeps working.
func (s *StringSlice) UnmarshalJSON(text []byte) error {
	var list []string
	if err := json.Unmarshal(text, &list); err == nil {
		*s = list
		return nil
	}
	data, err := strconv.Unquote(string(text))
	if err != nil {
		return fmt.Errorf("string slice must be a quoted string or an array, got %s: %w", text, err)
	}
	if data == "" {
		*s = nil
		return nil
	}
	*s = strings.Split(data, ",")
	return nil
}

// StoresInfo is the payload of GET /pd/api/v1/stores.
type StoresInfo struct {
	Count  int          `json:"count"`
	Stores []*StoreInfo `json:"stores"`
}

// StoreInfo contains information about a store.
type StoreInfo struct {
	Store  *MetaStore   `json:"store"`
	Status *StoreStatus `json:"status"`
}

// MetaStore contains meta information about a store.
type MetaStore struct {
	*Store
	StateName string `json:"state_name"`
}

// Store is the store meta information (the metapb.Store subset ticheck needs).
type Store struct {
	ID      uint64        `json:"id"`
	Address string        `json:"address"`
	State   int32         `json:"state"`
	Version string        `json:"version"`
	Labels  []*StoreLabel `json:"labels,omitempty"`
}

// StoreLabel is one label of a store, e.g. {key: "zone", value: "z1"}.
type StoreLabel struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// StoreStatus contains status about a store.
type StoreStatus struct {
	LeaderCount int `json:"leader_count"`
	RegionCount int `json:"region_count"`
}

// Config is the payload of GET /pd/api/v1/config, reduced to the sections
// dr-check reads.
type Config struct {
	Replication     ReplicationConfig     `json:"replication"`
	ReplicationMode ReplicationModeConfig `json:"replication-mode"`
}

// ReplicationConfig is the replication configuration.
type ReplicationConfig struct {
	// LocationLabels are the label keys that describe where a store lives.
	// Their order implies the isolation priority.
	LocationLabels StringSlice `json:"location-labels"`
}

// ReplicationModeConfig is the configuration for the replication policy.
type ReplicationModeConfig struct {
	// can be 'dr-auto-sync' or 'majority'
	ReplicationMode string                      `json:"replication-mode"`
	DRAutoSync      DRAutoSyncReplicationConfig `json:"dr-auto-sync"`
}

// DRAutoSyncReplicationConfig is the configuration for auto sync mode between
// two data centers.
type DRAutoSyncReplicationConfig struct {
	LabelKey         string   `json:"label-key"`
	Primary          string   `json:"primary"`
	DR               string   `json:"dr"`
	WaitStoreTimeout Duration `json:"wait-store-timeout"`
}

// HTTPReplicationStatus is the payload of
// GET /pd/api/v1/replication_mode/status.
type HTTPReplicationStatus struct {
	Mode       string `json:"mode"`
	DrAutoSync struct {
		LabelKey        string  `json:"label_key"`
		State           string  `json:"state"`
		StateID         uint64  `json:"state_id,omitempty"`
		ACIDConsistent  bool    `json:"acid_consistent"`
		TotalRegions    int     `json:"total_regions,omitempty"`
		SyncedRegions   int     `json:"synced_regions,omitempty"`
		RecoverProgress float32 `json:"recover_progress,omitempty"`
	} `json:"dr-auto-sync,omitempty"`
}

// Member is one PD member as returned by GET /pd/api/v1/members.
type Member struct {
	Name           string   `json:"name,omitempty"`
	MemberId       uint64   `json:"member_id,omitempty"`
	PeerUrls       []string `json:"peer_urls,omitempty"`
	ClientUrls     []string `json:"client_urls,omitempty"`
	LeaderPriority int32    `json:"leader_priority,omitempty"`
	DeployPath     string   `json:"deploy_path,omitempty"`
	BinaryVersion  string   `json:"binary_version,omitempty"`
	GitHash        string   `json:"git_hash,omitempty"`
	DcLocation     string   `json:"dc_location,omitempty"`
}

// GetMembersResponse is the payload of GET /pd/api/v1/members.
type GetMembersResponse struct {
	Members []*Member `json:"members,omitempty"`
	Leader  *Member   `json:"leader,omitempty"`
}

// Rule is the placement rule that can be checked against a region. When
// applying rules the order is defined by the tuple [GroupIndex, GroupID,
// Index, ID].
type Rule struct {
	GroupID          string            `json:"group_id"`                    // mark the source that add the rule
	ID               string            `json:"id"`                          // unique ID within a group
	Index            int               `json:"index,omitempty"`             // rule apply order in a group
	Role             PeerRoleType      `json:"role"`                        // expected role of the peers
	Count            int               `json:"count"`                       // expected count of the peers
	LabelConstraints []LabelConstraint `json:"label_constraints,omitempty"` // used to select stores to place peers
	LocationLabels   []string          `json:"location_labels,omitempty"`   // used to make peers isolated physically
}

// PeerRoleType is the expected peer type of the placement rule.
type PeerRoleType string

// LabelConstraint is used to filter stores when trying to place a peer.
type LabelConstraint struct {
	Key    string            `json:"key,omitempty"`
	Op     LabelConstraintOp `json:"op,omitempty"`
	Values []string          `json:"values,omitempty"`
}

// LabelConstraintOp defines how a LabelConstraint matches a store. It can be
// one of 'in', 'notIn', 'exists', or 'notExists'.
type LabelConstraintOp string

const (
	// In restricts the store label value should in the value list.
	// If label does not exist, `in` is always false.
	In LabelConstraintOp = "in"
	// NotIn restricts the store label value should not in the value list.
	// If label does not exist, `notIn` is always true.
	NotIn LabelConstraintOp = "notIn"
	// Exists restricts the store should have the label.
	Exists LabelConstraintOp = "exists"
	// NotExists restricts the store should not have the label.
	NotExists LabelConstraintOp = "notExists"
)

// MatchStore checks if a store matches the constraint.
func (c *LabelConstraint) MatchStore(store *StoreInfo) bool {
	switch c.Op {
	case In:
		label := store.GetLabelValue(c.Key)
		return label != "" && AnyOf(c.Values, func(i int) bool { return c.Values[i] == label })
	case NotIn:
		label := store.GetLabelValue(c.Key)
		return label == "" || NoneOf(c.Values, func(i int) bool { return c.Values[i] == label })
	case Exists:
		return store.GetLabelValue(c.Key) != ""
	case NotExists:
		return store.GetLabelValue(c.Key) == ""
	}
	return false
}

// GetLabelValue returns a label's value (if exists).
func (s *StoreInfo) GetLabelValue(key string) string {
	if s == nil || s.Store == nil {
		return ""
	}
	for _, label := range s.Store.Labels {
		if label != nil && strings.EqualFold(label.Key, key) {
			return label.Value
		}
	}
	return ""
}

// Capture holds common information of a capture in cdc.
type Capture struct {
	ID            string `json:"id"`
	IsOwner       bool   `json:"is_owner"`
	AdvertiseAddr string `json:"address"`
	ClusterID     string `json:"cluster_id"`
}

// ChangefeedCommonInfo holds some common usage information of a changefeed.
type ChangefeedCommonInfo struct {
	UpstreamID     uint64        `json:"upstream_id"`
	Namespace      string        `json:"namespace"`
	ID             string        `json:"id"`
	FeedState      string        `json:"state"`
	CheckpointTSO  uint64        `json:"checkpoint_tso"`
	CheckpointTime string        `json:"checkpoint_time"`
	RunningError   *RunningError `json:"error"`
}

// RunningError represents some running error from cdc components, such as
// processor.
type RunningError struct {
	Time    time.Time `json:"time"`
	Addr    string    `json:"addr"`
	Code    string    `json:"code"`
	Message string    `json:"message"`
}

// ListResponse is the response for all TiCDC List APIs.
type ListResponse[T any] struct {
	Total int `json:"total"`
	Items []T `json:"items"`
}
