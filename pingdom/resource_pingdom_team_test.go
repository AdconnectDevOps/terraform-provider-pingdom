package pingdom

import "testing"

func TestTeamForResource(t *testing.T) {
	d := resourcePingdomTeam().TestResourceData()
	d.Set("name", "platform-oncall")
	d.Set("member_ids", []interface{}{101, 202, 303})

	got, err := teamForResource(d)
	if err != nil {
		t.Fatalf("teamForResource: %v", err)
	}
	if got.Name != "platform-oncall" {
		t.Errorf("Name = %q, want platform-oncall", got.Name)
	}
	if len(got.MemberIDs) != 3 {
		t.Errorf("MemberIDs = %v, want 3 entries", got.MemberIDs)
	}

	want := map[int]bool{101: true, 202: true, 303: true}
	for _, id := range got.MemberIDs {
		if !want[id] {
			t.Errorf("unexpected MemberID %d in %v", id, got.MemberIDs)
		}
		delete(want, id)
	}
	if len(want) != 0 {
		t.Errorf("missing MemberIDs: %v", want)
	}
}

func TestTeamForResource_NoMembers(t *testing.T) {
	d := resourcePingdomTeam().TestResourceData()
	d.Set("name", "empty")

	got, err := teamForResource(d)
	if err != nil {
		t.Fatalf("teamForResource: %v", err)
	}
	if got.Name != "empty" {
		t.Errorf("Name = %q, want empty", got.Name)
	}
	if len(got.MemberIDs) != 0 {
		t.Errorf("MemberIDs = %v, want empty", got.MemberIDs)
	}
}
