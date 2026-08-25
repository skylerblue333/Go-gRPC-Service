package projects

import "testing"

func TestProjectLifecycleAndMembership(t *testing.T) {
	r := NewRegistry()
	created, err := r.Create("proj.alpha", "owner.1", "Alpha Project")
	if err != nil {
		t.Fatal(err)
	}
	if created.Version != 1 || created.Status != StatusPlanned || created.TaskExecutionPerformed {
		t.Fatalf("unexpected created snapshot: %+v", created)
	}

	withManager, err := r.AddMember("proj.alpha", "owner.1", "manager.1", RoleManager, created.Version)
	if err != nil {
		t.Fatal(err)
	}
	if withManager.Version != 2 {
		t.Fatalf("expected version 2, got %d", withManager.Version)
	}

	active, err := r.SetStatus("proj.alpha", "manager.1", StatusActive, withManager.Version)
	if err != nil {
		t.Fatal(err)
	}
	if active.Status != StatusActive || active.Version != 3 {
		t.Fatalf("unexpected active snapshot: %+v", active)
	}

	withTask, err := r.AddTaskRef("proj.alpha", "manager.1", "task.external-1", active.Version)
	if err != nil {
		t.Fatal(err)
	}
	if len(withTask.TaskRefs) != 1 || withTask.TaskRefs[0] != "task.external-1" {
		t.Fatalf("unexpected task refs: %+v", withTask.TaskRefs)
	}
	if withTask.TaskExecutionPerformed {
		t.Fatal("project registry must not claim task execution")
	}
}

func TestProjectRejectsStaleAndUnauthorizedMutations(t *testing.T) {
	r := NewRegistry()
	created, err := r.Create("proj.safe", "owner.2", "Safe")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := r.SetStatus("proj.safe", "owner.2", StatusActive, 99); err == nil || err.Error() != "version_conflict" {
		t.Fatalf("expected version_conflict, got %v", err)
	}

	member, err := r.AddMember("proj.safe", "owner.2", "member.1", RoleMember, created.Version)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.SetStatus("proj.safe", "member.1", StatusCompleted, member.Version); err == nil || err.Error() != "manager_required" {
		t.Fatalf("expected manager_required, got %v", err)
	}
	if _, err := r.AddMember("proj.safe", "member.1", "member.2", RoleMember, member.Version); err == nil || err.Error() != "owner_required" {
		t.Fatalf("expected owner_required, got %v", err)
	}
}

func TestProjectSnapshotsAreDeterministicAndDetached(t *testing.T) {
	r := NewRegistry()
	created, err := r.Create("proj.detached", "z.owner", "Detached")
	if err != nil {
		t.Fatal(err)
	}
	p, err := r.AddMember("proj.detached", "z.owner", "a.member", RoleMember, created.Version)
	if err != nil {
		t.Fatal(err)
	}
	p.Members[0].Role = RoleOwner
	p.TaskRefs = append(p.TaskRefs, "fake")

	fresh, err := r.Get("proj.detached")
	if err != nil {
		t.Fatal(err)
	}
	if fresh.Members[0].UserID != "a.member" || fresh.Members[0].Role != RoleMember {
		t.Fatalf("snapshot not deterministic/detached: %+v", fresh.Members)
	}
	if len(fresh.TaskRefs) != 0 {
		t.Fatalf("caller mutated retained task refs: %+v", fresh.TaskRefs)
	}
}
