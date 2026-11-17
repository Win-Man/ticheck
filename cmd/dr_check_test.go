package cmd

import (
	"reflect"
	"testing"

	"github.com/Win-Man/ticheck/pkg"
	"github.com/jedib0t/go-pretty/v6/table"
)

// newTestStore builds a pkg.StoreInfo as it would be decoded from
// /pd/api/v1/stores, so the table helpers can be exercised without a PD.
func newTestStore(address string, labels map[string]string, regionCount, leaderCount int) *pkg.StoreInfo {
	storeLabels := make([]*pkg.StoreLabel, 0, len(labels))
	for k, v := range labels {
		storeLabels = append(storeLabels, &pkg.StoreLabel{Key: k, Value: v})
	}
	return &pkg.StoreInfo{
		Store: &pkg.MetaStore{
			Store: &pkg.Store{
				Address: address,
				Labels:  storeLabels,
			},
		},
		Status: &pkg.StoreStatus{
			RegionCount: regionCount,
			LeaderCount: leaderCount,
		},
	}
}

// TestBuildStoreRowsThreeLocationLabels covers the topology the tool was
// written for: location-labels = ["zone", "rack", "host"].
func TestBuildStoreRowsThreeLocationLabels(t *testing.T) {
	locationLabels := []string{"zone", "rack", "host"}
	stores := []*pkg.StoreInfo{
		newTestStore("10.0.0.3:20160", map[string]string{"zone": "z2", "rack": "r1", "host": "h1"}, 300, 30),
		newTestStore("10.0.0.2:20160", map[string]string{"zone": "z1", "rack": "r2", "host": "h2"}, 200, 20),
		newTestStore("10.0.0.1:20160", map[string]string{"zone": "z1", "rack": "r1", "host": "h3"}, 100, 10),
	}

	rows := buildStoreRows(stores, locationLabels, nil)

	want := []table.Row{
		{"z1", "r1", "h3", "10.0.0.1:20160", pkg.PeerRoleType(""), 100, 10},
		{"z1", "r2", "h2", "10.0.0.2:20160", pkg.PeerRoleType(""), 200, 20},
		{"z2", "r1", "h1", "10.0.0.3:20160", pkg.PeerRoleType(""), 300, 30},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("buildStoreRows() mismatch\n got: %v\nwant: %v", rows, want)
	}
}

// TestBuildStoreRowsTwoLocationLabels covers the 同城两中心 topology where
// location-labels = ["zone", "dc"]. The old sorter compared columns 1..3, which
// here means dc / instance / role: it never touched zone, so rows were ordered
// by address rather than by zone.
func TestBuildStoreRowsTwoLocationLabels(t *testing.T) {
	locationLabels := []string{"zone", "dc"}
	stores := []*pkg.StoreInfo{
		newTestStore("10.0.0.2:20160", map[string]string{"zone": "z1", "dc": "dc2"}, 200, 20),
		newTestStore("10.0.0.1:20160", map[string]string{"zone": "z1", "dc": "dc1"}, 100, 10),
	}

	rows := buildStoreRows(stores, locationLabels, nil)

	want := []table.Row{
		{"z1", "dc1", "10.0.0.1:20160", pkg.PeerRoleType(""), 100, 10},
		{"z1", "dc2", "10.0.0.2:20160", pkg.PeerRoleType(""), 200, 20},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("buildStoreRows() mismatch\n got: %v\nwant: %v", rows, want)
	}
}

// TestBuildStoreRowsManyLocationLabels guards the opposite side: with more
// labels than the old hard coded three, every label column must still render.
func TestBuildStoreRowsManyLocationLabels(t *testing.T) {
	locationLabels := []string{"region", "zone", "rack", "host"}
	stores := []*pkg.StoreInfo{
		newTestStore("10.0.0.1:20160",
			map[string]string{"region": "cn-north", "zone": "z1", "rack": "r1", "host": "h1"}, 10, 1),
	}

	rows := buildStoreRows(stores, locationLabels, nil)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if len(rows[0]) != 4+4 {
		t.Fatalf("expected %d columns, got %d: %v", 4+4, len(rows[0]), rows[0])
	}
	want := table.Row{"cn-north", "z1", "r1", "h1", "10.0.0.1:20160", pkg.PeerRoleType(""), 10, 1}
	if !reflect.DeepEqual(rows[0], want) {
		t.Fatalf("row mismatch\n got: %v\nwant: %v", rows[0], want)
	}
}

// TestBuildStoreRowsNoLocationLabels is the direct regression test for the old
// sorter panic: with no location-labels the columns are
// [instance, role, region, leader], and the hard coded loop started at index 1
// — the Role column — whose value is pkg.PeerRoleType. The string assertion
// there panicked with "interface {} is pkg.PeerRoleType, not string".
// An empty location-labels list is legal: PD only fills it in when the cluster
// still relies on max-replicas instead of placement rules.
func TestBuildStoreRowsNoLocationLabels(t *testing.T) {
	stores := []*pkg.StoreInfo{
		newTestStore("10.0.0.2:20160", map[string]string{"zone": "z2"}, 200, 20),
		newTestStore("10.0.0.1:20160", map[string]string{"zone": "z1"}, 100, 10),
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("buildStoreRows() panicked with no location labels: %v", r)
		}
	}()

	rows := buildStoreRows(stores, nil, nil)

	want := []table.Row{
		{"10.0.0.2:20160", pkg.PeerRoleType(""), 200, 20},
		{"10.0.0.1:20160", pkg.PeerRoleType(""), 100, 10},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("buildStoreRows() mismatch\n got: %v\nwant: %v", rows, want)
	}
}

// TestBuildStoreRowsSortsByFirstLabel is the regression test for the old sorter
// skipping column 0: with LocationLabels = [zone, dc, host] it compared
// dc / host / instance, so two stores from different zones were ordered by
// address instead of by zone. Cluster IPs rarely line up with zones, so this
// showed up as a wrong primary/DR grouping in the rendered table.
func TestBuildStoreRowsSortsByFirstLabel(t *testing.T) {
	locationLabels := []string{"zone", "dc", "host"}
	stores := []*pkg.StoreInfo{
		// z2 owns the smaller address on purpose
		newTestStore("10.0.0.1:20160", map[string]string{"zone": "z2", "dc": "dc1", "host": "h1"}, 200, 20),
		newTestStore("10.0.0.2:20160", map[string]string{"zone": "z1", "dc": "dc1", "host": "h2"}, 100, 10),
	}

	rows := buildStoreRows(stores, locationLabels, nil)

	if rows[0][0] != "z1" || rows[1][0] != "z2" {
		t.Fatalf("rows not sorted by the first location label: %v", rows)
	}
}

// TestBuildStoreRowsSkipsTiFlash keeps the original behaviour: TiFlash nodes are
// not part of the dr-auto-sync replica layout.
func TestBuildStoreRowsSkipsTiFlash(t *testing.T) {
	stores := []*pkg.StoreInfo{
		newTestStore("10.0.0.1:20160", map[string]string{"zone": "z1", "engine": "tikv"}, 100, 10),
		newTestStore("10.0.0.2:3930", map[string]string{"zone": "z1", "engine": "tiflash"}, 50, 0),
	}

	rows := buildStoreRows(stores, []string{"zone"}, nil)
	if len(rows) != 1 {
		t.Fatalf("expected only the tikv store, got %d rows: %v", len(rows), rows)
	}
	if rows[0][1] != "10.0.0.1:20160" {
		t.Fatalf("expected tikv store in row, got %v", rows[0])
	}
}

// TestBuildStoreRowsMissingLabel checks that a store missing one of the
// configured labels renders an empty cell instead of shifting the columns.
func TestBuildStoreRowsMissingLabel(t *testing.T) {
	stores := []*pkg.StoreInfo{
		newTestStore("10.0.0.1:20160", map[string]string{"zone": "z1"}, 100, 10),
	}

	rows := buildStoreRows(stores, []string{"zone", "rack", "host"}, nil)
	want := []table.Row{
		{"z1", "", "", "10.0.0.1:20160", pkg.PeerRoleType(""), 100, 10},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("buildStoreRows() mismatch\n got: %v\nwant: %v", rows, want)
	}
}

// TestBuildStoreRowsNilSafety makes sure malformed API payloads degrade to
// skipped rows instead of a nil pointer dereference.
func TestBuildStoreRowsNilSafety(t *testing.T) {
	stores := []*pkg.StoreInfo{
		nil,
		{Store: nil},
		{Store: &pkg.MetaStore{Store: &pkg.Store{Address: "10.0.0.9:20160"}}, Status: nil},
	}

	rows := buildStoreRows(stores, []string{"zone"}, nil)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d: %v", len(rows), rows)
	}
	want := table.Row{"", "10.0.0.9:20160", pkg.PeerRoleType(""), 0, 0}
	if !reflect.DeepEqual(rows[0], want) {
		t.Fatalf("row mismatch\n got: %v\nwant: %v", rows[0], want)
	}
}

// TestMatchStoreRole documents the current rule resolution: the role of the
// last matching placement rule wins. Changing this to PD's rule-index ordering
// must update this test on purpose.
func TestMatchStoreRole(t *testing.T) {
	rules := []pkg.Rule{
		{
			GroupID: "pd",
			ID:      "primary",
			Role:    pkg.PeerRoleType("voter"),
			LabelConstraints: []pkg.LabelConstraint{
				{Key: "zone", Op: pkg.In, Values: []string{"z1"}},
			},
		},
		{
			GroupID: "pd",
			ID:      "dr",
			Role:    pkg.PeerRoleType("learner"),
			LabelConstraints: []pkg.LabelConstraint{
				{Key: "dc", Op: pkg.In, Values: []string{"dc2"}},
			},
		},
	}

	tests := []struct {
		name   string
		labels map[string]string
		want   pkg.PeerRoleType
	}{
		{"matches first rule only", map[string]string{"zone": "z1", "dc": "dc1"}, "voter"},
		{"matches second rule only", map[string]string{"zone": "z9", "dc": "dc2"}, "learner"},
		{"matches both, last wins", map[string]string{"zone": "z1", "dc": "dc2"}, "learner"},
		{"matches nothing", map[string]string{"zone": "z9", "dc": "dc9"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore("10.0.0.1:20160", tt.labels, 0, 0)
			if got := matchStoreRole(store, rules); got != tt.want {
				t.Errorf("matchStoreRole() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMatchStoreRoleNilStore(t *testing.T) {
	rules := []pkg.Rule{{Role: pkg.PeerRoleType("voter")}}
	if got := matchStoreRole(nil, rules); got != "" {
		t.Fatalf("matchStoreRole(nil) = %q, want empty", got)
	}
}

// TestSortRowsByLabels checks mixed column types sort without a type assertion
// panic and that the sort is stable for identical label prefixes.
func TestSortRowsByLabels(t *testing.T) {
	rows := []table.Row{
		{"z2", "r1", "10.0.0.3:20160", pkg.PeerRoleType("voter"), 300, 30},
		{"z1", "r2", "10.0.0.2:20160", pkg.PeerRoleType("learner"), 200, 20},
		{"z1", "r1", "10.0.0.1:20160", pkg.PeerRoleType("voter"), 100, 10},
	}

	sortRowsByLabels(rows, 2)

	want := []table.Row{
		{"z1", "r1", "10.0.0.1:20160", pkg.PeerRoleType("voter"), 100, 10},
		{"z1", "r2", "10.0.0.2:20160", pkg.PeerRoleType("learner"), 200, 20},
		{"z2", "r1", "10.0.0.3:20160", pkg.PeerRoleType("voter"), 300, 30},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("sortRowsByLabels() mismatch\n got: %v\nwant: %v", rows, want)
	}
}

func TestSortRowsByLabelsIgnoresExtraColumns(t *testing.T) {
	rows := []table.Row{
		{"z1", "10.0.0.2:20160"},
		{"z1", "10.0.0.1:20160"},
	}

	// sorting by 0 label columns must keep the original order
	sortRowsByLabels(rows, 0)
	if rows[0][1] != "10.0.0.2:20160" {
		t.Fatalf("expected original order to be preserved, got %v", rows)
	}
}
