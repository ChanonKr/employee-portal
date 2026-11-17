package entity

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
	Skills       []Skill
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
		return ErrConflict
	}
	return nil
}

// ChangeContact applies one contact edit. The caller must have loaded e under a row
// lock. It returns the Change to record, or nil when the value did not change
// (the version still increments on save, but no history entry is written).
func (e *Employee) ChangeContact(f ContactField, raw string, expectedVersion int32) (*Change, error) {
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
	return &Change{Field: f.LogField(), Old: old, New: v}, nil
}
