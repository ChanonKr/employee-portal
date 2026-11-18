// Package domain holds the profile rules: contact fields, skills selection, photo,
// legal name and optimistic locking. No framework imports.
package domain

import (
	historydomain "employee-portal/modules/history/domain"
	"employee-portal/modules/shared"
	skilldomain "employee-portal/modules/skill/domain"
)

// Employee is the editable profile. The name is read-only for employees.
type Employee struct {
	ID           int64
	UserID       int64
	FirstName    string
	LastName     string
	PhotoPath    *string
	MobilePhone  *string
	ContactEmail *string
	Address      *string
	Version      int32 // optimistic locking: every write increments it
}

// Profile is what the employee sees on the profile page.
type Profile struct {
	Employee
	EmployeeCode string
	Skills       []skilldomain.Skill
}

// Contact returns the current value of a contact field.
func (e *Employee) Contact(f ContactField) *string {
	switch f {
	case ContactMobile:
		return e.MobilePhone
	case ContactEmail:
		return e.ContactEmail
	default:
		return e.Address
	}
}

func (e *Employee) setContact(f ContactField, v *string) {
	switch f {
	case ContactMobile:
		e.MobilePhone = v
	case ContactEmail:
		e.ContactEmail = v
	default:
		e.Address = v
	}
}

// CheckVersion enforces optimistic locking: the client must have seen the latest version.
func (e *Employee) CheckVersion(expected int32) error {
	if e.Version != expected {
		return shared.ErrConflict
	}
	return nil
}

// ChangeContact applies one contact edit. The caller must have loaded e under a row
// lock. It returns the Change to record, or nil when the value did not change
// (the version still increments on save, but no history entry is written).
func (e *Employee) ChangeContact(f ContactField, raw string, expectedVersion int32) (*historydomain.Change, error) {
	v := f.Normalize(raw)
	if err := f.Validate(v); err != nil {
		return nil, err
	}
	if err := e.CheckVersion(expectedVersion); err != nil {
		return nil, err
	}

	current := e.Contact(f)
	e.setContact(f, &v)
	if current != nil && *current == v {
		return nil, nil
	}

	var old any // keep nil as an untyped nil, not a nil *string
	if current != nil {
		old = *current
	}
	return &historydomain.Change{Field: f.LogField(), Old: old, New: v}, nil
}

// ChangeSkills checks a request to replace the employee's whole skill set.
// Skills live outside the employees row, so the caller loads them under the row
// lock and passes them in:
//   - current:  the skills the employee has now
//   - selected: the catalog entries for the requested ids (active or not)
//
// It returns the Change to record, or nil when the set is the same
// (the version still increments on save, but no history entry is written).
func (e *Employee) ChangeSkills(requested []int64, current, selected []skilldomain.Skill, expectedVersion int32) (*historydomain.Change, error) {
	if err := skilldomain.ValidateSkillIDs(requested); err != nil {
		return nil, err
	}
	if err := e.CheckVersion(expectedVersion); err != nil {
		return nil, err
	}
	if err := skilldomain.CheckSelection(requested, selected, current); err != nil {
		return nil, err
	}
	if skilldomain.SameSet(current, selected) {
		return nil, nil
	}
	return &historydomain.Change{
		Field: historydomain.FieldSkills,
		Old:   skilldomain.Names(current),
		New:   skilldomain.Names(selected),
	}, nil
}

// ChangePhoto replaces the photo with a newly stored file. The caller must have
// loaded e under a row lock. Old is the previous file name (nil if none) so the
// caller can delete it after commit.
func (e *Employee) ChangePhoto(newPath string, expectedVersion int32) (*historydomain.Change, error) {
	if err := e.CheckVersion(expectedVersion); err != nil {
		return nil, err
	}
	var old any
	if e.PhotoPath != nil {
		old = *e.PhotoPath
	}
	e.PhotoPath = &newPath
	return &historydomain.Change{Field: historydomain.FieldPhoto, Old: old, New: newPath}, nil
}
