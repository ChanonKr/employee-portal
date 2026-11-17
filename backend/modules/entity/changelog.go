package entity

// FieldName is the profile field recorded in the change history.
type FieldName string

const (
	FieldFirstName    FieldName = "first_name"
	FieldLastName     FieldName = "last_name"
	FieldMobilePhone  FieldName = "mobile_phone"
	FieldContactEmail FieldName = "contact_email"
	FieldAddress      FieldName = "address"
	FieldPhoto        FieldName = "photo"
	FieldSkills       FieldName = "skills"
)

// Change is one field edit. Old and New are nil, a string, or []string (skills).
type Change struct {
	Field FieldName
	Old   any
	New   any
}
