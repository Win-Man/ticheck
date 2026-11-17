package pkg

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// The fixtures below are trimmed copies of real PD / TiCDC responses. They keep
// the fields ticheck models plus a handful it does not (capacity, max-replicas,
// cluster-version, ...) so the tests fail if trimming the structs ever breaks
// decoding of a response whose payload still carries those keys.

const pdConfigFixture = `{
  "client-urls": "http://0.0.0.0:2379",
  "peer-urls": "http://0.0.0.0:2380",
  "lease": 3,
  "tso-save-interval": "1m0s",
  "schedule": {"max-snapshot-count": 64, "max-store-down-time": "30m0s"},
  "replication": {
    "max-replicas": 3,
    "location-labels": "zone,dc,host",
    "strictly-match-label": "false",
    "enable-placement-rules": "true"
  },
  "pd-server": {"max-gap-reset-ts": "24h0m0s", "runtime-services": "tso"},
  "cluster-version": "v5.4.0",
  "quota-backend-bytes": "8GiB",
  "replication-mode": {
    "replication-mode": "dr-auto-sync",
    "dr-auto-sync": {
      "label-key": "dc",
      "primary": "dc1",
      "dr": "dc2",
      "primary-replicas": 2,
      "dr-replicas": 1,
      "wait-store-timeout": "1m0s"
    }
  },
  "label-property": {"zone": [{"key": "reject-leader", "value": "true"}]}
}`

func TestDecodePDConfig(t *testing.T) {
	var cfg Config
	if err := json.Unmarshal([]byte(pdConfigFixture), &cfg); err != nil {
		t.Fatalf("decode pd config failed: %v", err)
	}

	wantLabels := StringSlice{"zone", "dc", "host"}
	if !reflect.DeepEqual(cfg.Replication.LocationLabels, wantLabels) {
		t.Fatalf("location-labels\n got: %#v\nwant: %#v", cfg.Replication.LocationLabels, wantLabels)
	}

	if got := cfg.ReplicationMode.ReplicationMode; got != "dr-auto-sync" {
		t.Fatalf("replication-mode = %q, want dr-auto-sync", got)
	}
	dr := cfg.ReplicationMode.DRAutoSync
	if dr.LabelKey != "dc" || dr.Primary != "dc1" || dr.DR != "dc2" {
		t.Fatalf("unexpected dr-auto-sync config: %+v", dr)
	}
	if dr.WaitStoreTimeout.Duration != time.Minute {
		t.Fatalf("wait-store-timeout = %v, want 1m", dr.WaitStoreTimeout)
	}
	// %v on the struct goes through the embedded time.Duration's String, which
	// is what dr-check prints.
	if got := dr.WaitStoreTimeout.String(); got != "1m0s" {
		t.Fatalf("wait-store-timeout String() = %q, want 1m0s", got)
	}
}

// TestDecodePDConfigWithoutLocationLabels covers the drift the tool has to
// survive: PD sends an empty string when location-labels is not configured.
func TestDecodePDConfigWithoutLocationLabels(t *testing.T) {
	body := `{"replication":{"location-labels":""},"replication-mode":{"replication-mode":"majority"}}`

	var cfg Config
	if err := json.Unmarshal([]byte(body), &cfg); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(cfg.Replication.LocationLabels) != 0 {
		t.Fatalf("expected no location labels, got %#v", cfg.Replication.LocationLabels)
	}
	if cfg.ReplicationMode.ReplicationMode != "majority" {
		t.Fatalf("replication-mode = %q, want majority", cfg.ReplicationMode.ReplicationMode)
	}
}

func TestStringSliceAcceptsArrayForm(t *testing.T) {
	body := `{"replication":{"location-labels":["zone","host"]}}`

	var cfg Config
	if err := json.Unmarshal([]byte(body), &cfg); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	want := StringSlice{"zone", "host"}
	if !reflect.DeepEqual(cfg.Replication.LocationLabels, want) {
		t.Fatalf("location-labels\n got: %#v\nwant: %#v", cfg.Replication.LocationLabels, want)
	}
}

func TestStringSliceInvalidType(t *testing.T) {
	body := `{"replication":{"location-labels":42}}`

	var cfg Config
	if err := json.Unmarshal([]byte(body), &cfg); err == nil {
		t.Fatalf("expected an error for a numeric location-labels, got %#v", cfg.Replication.LocationLabels)
	}
}

func TestStringSliceMarshal(t *testing.T) {
	got, err := json.Marshal(StringSlice{"zone", "host"})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if string(got) != `"zone,host"` {
		t.Fatalf("marshal = %s, want \"zone,host\"", got)
	}

	empty, err := json.Marshal(StringSlice(nil))
	if err != nil {
		t.Fatalf("marshal nil failed: %v", err)
	}
	if string(empty) != `""` {
		t.Fatalf("marshal nil = %s, want \"\"", empty)
	}
}

func TestDurationRoundTrip(t *testing.T) {
	var d Duration
	if err := json.Unmarshal([]byte(`"30s"`), &d); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if d.Duration != 30*time.Second {
		t.Fatalf("decoded %v, want 30s", d.Duration)
	}

	out, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if string(out) != `"30s"` {
		t.Fatalf("encoded %s, want \"30s\"", out)
	}

	if err := json.Unmarshal([]byte(`"not-a-duration"`), &d); err == nil {
		t.Fatal("expected an error for an invalid duration")
	}
	if err := json.Unmarshal([]byte(`30`), &d); err == nil {
		t.Fatal("expected an error for an unquoted duration")
	}
}

const pdStoresFixture = `{
  "count": 2,
  "stores": [
    {
      "store": {
        "id": 1,
        "address": "10.0.0.1:20160",
        "state": 0,
        "version": "5.4.0",
        "labels": [{"key": "zone", "value": "z1"}, {"key": "dc", "value": "dc1"}, {"key": "engine", "value": "tiflash"}]
      },
      "status": {"capacity": "100GiB", "available": "50GiB", "leader_count": 10, "region_count": 100, "uptime": "1h0m0s"}
    },
    {
      "store": {"id": 2, "address": "10.0.0.2:20160", "state": 0, "version": "5.4.0"},
      "status": {"capacity": "100GiB", "available": "40GiB", "leader_count": 20, "region_count": 200}
    }
  ]
}`

func TestDecodeStoresInfo(t *testing.T) {
	var info StoresInfo
	if err := json.Unmarshal([]byte(pdStoresFixture), &info); err != nil {
		t.Fatalf("decode stores failed: %v", err)
	}

	if info.Count != 2 || len(info.Stores) != 2 {
		t.Fatalf("expected 2 stores, got count=%d len=%d", info.Count, len(info.Stores))
	}
	first := info.Stores[0]
	if first.Store == nil || first.Store.Address != "10.0.0.1:20160" {
		t.Fatalf("store address not decoded: %+v", first.Store)
	}
	if first.Status == nil || first.Status.RegionCount != 100 || first.Status.LeaderCount != 10 {
		t.Fatalf("store status not decoded: %+v", first.Status)
	}
	if got := first.GetLabelValue("zone"); got != "z1" {
		t.Fatalf("GetLabelValue(zone) = %q, want z1", got)
	}
	if got := first.GetLabelValue("DC"); got != "dc1" {
		t.Fatalf("GetLabelValue is case insensitive, got %q for DC", got)
	}
	if got := first.GetLabelValue("rack"); got != "" {
		t.Fatalf("missing label should be empty, got %q", got)
	}

	// a store without labels must not panic or invent values
	second := info.Stores[1]
	if got := second.GetLabelValue("zone"); got != "" {
		t.Fatalf("store without labels returned %q", got)
	}
}

// TestLabelConstraintMatchStore pins the matching rules dr-check relies on when
// it derives a store's role from the placement rules.
func TestLabelConstraintMatchStore(t *testing.T) {
	store := &StoreInfo{
		Store: &MetaStore{Store: &Store{
			Address: "10.0.0.1:20160",
			Labels: []*StoreLabel{
				{Key: "dc", Value: "dc1"},
				{Key: "zone", Value: "z1"},
			},
		}},
	}

	cases := []struct {
		name       string
		constraint LabelConstraint
		want       bool
	}{
		{"in hits", LabelConstraint{Key: "dc", Op: In, Values: []string{"dc1", "dc2"}}, true},
		{"in misses", LabelConstraint{Key: "dc", Op: In, Values: []string{"dc2"}}, false},
		{"in on missing label", LabelConstraint{Key: "rack", Op: In, Values: []string{"r1"}}, false},
		{"notIn misses", LabelConstraint{Key: "dc", Op: NotIn, Values: []string{"dc1"}}, false},
		{"notIn hits", LabelConstraint{Key: "dc", Op: NotIn, Values: []string{"dc2"}}, true},
		{"notIn on missing label", LabelConstraint{Key: "rack", Op: NotIn, Values: []string{"r1"}}, true},
		{"exists", LabelConstraint{Key: "zone", Op: Exists}, true},
		{"exists on missing label", LabelConstraint{Key: "rack", Op: Exists}, false},
		{"notExists", LabelConstraint{Key: "rack", Op: NotExists}, true},
		{"notExists on existing label", LabelConstraint{Key: "zone", Op: NotExists}, false},
		{"unknown op", LabelConstraint{Key: "zone", Op: LabelConstraintOp("bogus")}, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.constraint.MatchStore(store); got != c.want {
				t.Fatalf("MatchStore() = %v, want %v", got, c.want)
			}
		})
	}
}

const pdRulesFixture = `[
  {
    "group_id": "pd",
    "id": "dc1",
    "index": 1,
    "role": "voter",
    "count": 2,
    "label_constraints": [{"key": "dc", "op": "in", "values": ["dc1"]}],
    "location_labels": ["zone", "host"]
  },
  {
    "group_id": "pd",
    "id": "dc2",
    "index": 2,
    "role": "voter",
    "count": 1,
    "label_constraints": [{"key": "dc", "op": "in", "values": ["dc2"]}],
    "location_labels": ["zone", "host"]
  }
]`

func TestDecodePlacementRules(t *testing.T) {
	var rules []Rule
	if err := json.Unmarshal([]byte(pdRulesFixture), &rules); err != nil {
		t.Fatalf("decode rules failed: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[0].Role != PeerRoleType("voter") || rules[0].Count != 2 {
		t.Fatalf("rule 0 not decoded: %+v", rules[0])
	}
	if len(rules[0].LabelConstraints) != 1 || rules[0].LabelConstraints[0].Op != In {
		t.Fatalf("rule 0 constraints not decoded: %+v", rules[0].LabelConstraints)
	}

	dc1Store := &StoreInfo{Store: &MetaStore{Store: &Store{
		Labels: []*StoreLabel{{Key: "dc", Value: "dc1"}},
	}}}
	if !rules[0].LabelConstraints[0].MatchStore(dc1Store) {
		t.Fatal("dc1 store should match the dc1 rule")
	}
	if rules[1].LabelConstraints[0].MatchStore(dc1Store) {
		t.Fatal("dc1 store should not match the dc2 rule")
	}
}

func TestDecodeReplicationStatus(t *testing.T) {
	body := `{"mode":"dr-auto-sync","dr-auto-sync":{"label_key":"dc","state":"sync","state_id":5,"acid_consistent":true,"total_regions":100,"synced_regions":100}}`

	var status HTTPReplicationStatus
	if err := json.Unmarshal([]byte(body), &status); err != nil {
		t.Fatalf("decode status failed: %v", err)
	}
	if status.Mode != "dr-auto-sync" || status.DrAutoSync.State != "sync" {
		t.Fatalf("unexpected status: %+v", status)
	}
	if status.DrAutoSync.LabelKey != "dc" || status.DrAutoSync.TotalRegions != 100 {
		t.Fatalf("unexpected dr-auto-sync payload: %+v", status.DrAutoSync)
	}
}

func TestDecodeMembers(t *testing.T) {
	body := `{
      "header": {"cluster_id": 1234},
      "members": [
        {"name": "pd-1", "member_id": 1, "client_urls": ["http://10.0.0.1:2379"], "leader_priority": 0},
        {"name": "pd-2", "member_id": 2, "client_urls": ["http://10.0.0.2:2379"], "leader_priority": 5}
      ],
      "leader": {"name": "pd-1", "member_id": 1, "client_urls": ["http://10.0.0.1:2379"]}
    }`

	var resp GetMembersResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("decode members failed: %v", err)
	}
	if len(resp.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(resp.Members))
	}
	if resp.Leader == nil || resp.Leader.ClientUrls[0] != "http://10.0.0.1:2379" {
		t.Fatalf("leader not decoded: %+v", resp.Leader)
	}
	if resp.Members[1].LeaderPriority != 5 {
		t.Fatalf("leader_priority not decoded: %+v", resp.Members[1])
	}
}

func TestDecodeCDCLists(t *testing.T) {
	capturesBody := `{"total":1,"items":[{"id":"c1","is_owner":true,"address":"10.0.0.1:8300","cluster_id":"1234"}]}`
	var captures ListResponse[Capture]
	if err := json.Unmarshal([]byte(capturesBody), &captures); err != nil {
		t.Fatalf("decode captures failed: %v", err)
	}
	if captures.Total != 1 || captures.Items[0].AdvertiseAddr != "10.0.0.1:8300" || !captures.Items[0].IsOwner {
		t.Fatalf("unexpected captures: %+v", captures)
	}

	// shape of GET /api/v2/changefeeds on TiCDC v5.x
	changefeedsBody := `{"total":1,"items":[{
      "upstream_id": 7000000000000000000,
      "namespace": "default",
      "id": "test-cf",
      "state": "normal",
      "checkpoint_tso": 435932770427486209,
      "checkpoint_time": "2026-11-17 10:00:00.000",
      "error": null
    }]}`
	var changefeeds ListResponse[ChangefeedCommonInfo]
	if err := json.Unmarshal([]byte(changefeedsBody), &changefeeds); err != nil {
		t.Fatalf("decode changefeeds failed: %v", err)
	}
	cf := changefeeds.Items[0]
	if cf.ID != "test-cf" || cf.FeedState != "normal" || cf.RunningError != nil {
		t.Fatalf("unexpected changefeed: %+v", cf)
	}
}
