package entity

import "errors"

// Photo rules. The image itself is checked and normalized by pkg/imaging.
const (
	MaxPhotoSide = 512 // stored photos fit in 512×512
)

var (
	ErrPhotoTooLarge = errors.New("photo is too large")
	ErrPhotoFormat   = errors.New("photo must be a JPEG, PNG or WebP image")
)

// ChangePhoto replaces the photo with a newly stored file. The caller must have
// loaded e under a row lock. Old is the previous file name (nil if none) so the
// caller can delete it after commit.
func (e *Employee) ChangePhoto(newPath string, expectedVersion int32) (*Change, error) {
	if err := e.CheckVersion(expectedVersion); err != nil {
		return nil, err
	}
	var old any
	if e.PhotoPath != nil {
		old = *e.PhotoPath
	}
	e.PhotoPath = &newPath
	return &Change{Field: FieldPhoto, Old: old, New: newPath}, nil
}
