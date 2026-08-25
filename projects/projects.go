package projects

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type Status string

type Role string

const (
	StatusPlanned   Status = "planned"
	StatusActive    Status = "active"
	StatusPaused    Status = "paused"
	StatusCompleted Status = "completed"
	StatusArchived  Status = "archived"

	RoleOwner   Role = "owner"
	RoleManager Role = "manager"
	RoleMember  Role = "member"

	maxProjects = 10000
	maxMembers  = 1000
	maxTaskRefs = 5000
)

var safeID = regexp.MustCompile(`^[A-Za-z0-9_.:@-]{1,64}$`)

type Member struct {
	UserID string `json:"userId"`
	Role   Role   `json:"role"`
}

type Snapshot struct {
	ID                     string   `json:"id"`
	OwnerID                string   `json:"ownerId"`
	Name                   string   `json:"name"`
	Status                 Status   `json:"status"`
	Version                uint64   `json:"version"`
	Members                []Member `json:"members"`
	TaskRefs               []string `json:"taskRefs"`
	TaskExecutionPerformed bool     `json:"taskExecutionPerformed"`
}

type project struct {
	id       string
	ownerID  string
	name     string
	status   Status
	version  uint64
	members  map[string]Role
	taskRefs map[string]struct{}
}

type Registry struct {
	mu       sync.RWMutex
	projects map[string]*project
}

func NewRegistry() *Registry {
	return &Registry{projects: make(map[string]*project)}
}

func normalizeID(value, field string) (string, error) {
	value = strings.TrimSpace(value)
	if !safeID.MatchString(value) {
		return "", errors.New("invalid_" + field)
	}
	return value, nil
}

func normalizeName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if len(value) < 1 || len(value) > 160 {
		return "", errors.New("invalid_name")
	}
	return value, nil
}

func validStatus(status Status) bool {
	switch status {
	case StatusPlanned, StatusActive, StatusPaused, StatusCompleted, StatusArchived:
		return true
	default:
		return false
	}
}

func (r *Registry) Create(id, ownerID, name string) (Snapshot, error) {
	id, err := normalizeID(id, "project_id")
	if err != nil {
		return Snapshot{}, err
	}
	ownerID, err = normalizeID(ownerID, "owner_id")
	if err != nil {
		return Snapshot{}, err
	}
	name, err = normalizeName(name)
	if err != nil {
		return Snapshot{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.projects[id]; exists {
		return Snapshot{}, errors.New("project_exists")
	}
	if len(r.projects) >= maxProjects {
		return Snapshot{}, errors.New("capacity_exhausted")
	}
	p := &project{
		id:       id,
		ownerID:  ownerID,
		name:     name,
		status:   StatusPlanned,
		version:  1,
		members:  map[string]Role{ownerID: RoleOwner},
		taskRefs: make(map[string]struct{}),
	}
	r.projects[id] = p
	return snapshot(p), nil
}

func (r *Registry) Get(id string) (Snapshot, error) {
	id, err := normalizeID(id, "project_id")
	if err != nil {
		return Snapshot{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.projects[id]
	if !ok {
		return Snapshot{}, errors.New("project_not_found")
	}
	return snapshot(p), nil
}

func (r *Registry) SetStatus(id, actorID string, status Status, expectedVersion uint64) (Snapshot, error) {
	if !validStatus(status) {
		return Snapshot{}, errors.New("invalid_status")
	}
	return r.mutate(id, actorID, expectedVersion, func(p *project, actor Role) error {
		if actor != RoleOwner && actor != RoleManager {
			return errors.New("manager_required")
		}
		p.status = status
		return nil
	})
}

func (r *Registry) AddMember(id, actorID, userID string, role Role, expectedVersion uint64) (Snapshot, error) {
	if role != RoleManager && role != RoleMember {
		return Snapshot{}, errors.New("invalid_role")
	}
	userID, err := normalizeID(userID, "user_id")
	if err != nil {
		return Snapshot{}, err
	}
	return r.mutate(id, actorID, expectedVersion, func(p *project, actor Role) error {
		if actor != RoleOwner {
			return errors.New("owner_required")
		}
		if userID == p.ownerID {
			return errors.New("owner_role_immutable")
		}
		if _, exists := p.members[userID]; !exists && len(p.members) >= maxMembers {
			return errors.New("member_capacity_exhausted")
		}
		p.members[userID] = role
		return nil
	})
}

func (r *Registry) AddTaskRef(id, actorID, taskID string, expectedVersion uint64) (Snapshot, error) {
	taskID, err := normalizeID(taskID, "task_id")
	if err != nil {
		return Snapshot{}, err
	}
	return r.mutate(id, actorID, expectedVersion, func(p *project, actor Role) error {
		if actor != RoleOwner && actor != RoleManager {
			return errors.New("manager_required")
		}
		if _, exists := p.taskRefs[taskID]; !exists && len(p.taskRefs) >= maxTaskRefs {
			return errors.New("task_ref_capacity_exhausted")
		}
		p.taskRefs[taskID] = struct{}{}
		return nil
	})
}

func (r *Registry) mutate(id, actorID string, expectedVersion uint64, fn func(*project, Role) error) (Snapshot, error) {
	id, err := normalizeID(id, "project_id")
	if err != nil {
		return Snapshot{}, err
	}
	actorID, err = normalizeID(actorID, "actor_id")
	if err != nil {
		return Snapshot{}, err
	}
	if expectedVersion < 1 {
		return Snapshot{}, errors.New("invalid_expected_version")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.projects[id]
	if !ok {
		return Snapshot{}, errors.New("project_not_found")
	}
	if p.version != expectedVersion {
		return Snapshot{}, errors.New("version_conflict")
	}
	role, ok := p.members[actorID]
	if !ok {
		return Snapshot{}, errors.New("member_required")
	}
	if err := fn(p, role); err != nil {
		return Snapshot{}, err
	}
	p.version++
	return snapshot(p), nil
}

func snapshot(p *project) Snapshot {
	members := make([]Member, 0, len(p.members))
	for userID, role := range p.members {
		members = append(members, Member{UserID: userID, Role: role})
	}
	sort.Slice(members, func(i, j int) bool { return members[i].UserID < members[j].UserID })

	taskRefs := make([]string, 0, len(p.taskRefs))
	for id := range p.taskRefs {
		taskRefs = append(taskRefs, id)
	}
	sort.Strings(taskRefs)

	return Snapshot{
		ID:                     p.id,
		OwnerID:                p.ownerID,
		Name:                   p.name,
		Status:                 p.status,
		Version:                p.version,
		Members:                members,
		TaskRefs:               taskRefs,
		TaskExecutionPerformed: false,
	}
}
