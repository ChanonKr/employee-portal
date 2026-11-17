package httputil

import "path"

const photoURLPrefix = "/api/v1/files/photos/"

// PhotoURL turns a stored file name into the URL served by GET /files/photos/:name.
func PhotoURL(name string) string {
	return photoURLPrefix + path.Base(name)
}

// OptionalPhotoURL is PhotoURL for a nullable column.
func OptionalPhotoURL(name *string) *string {
	if name == nil {
		return nil
	}
	u := PhotoURL(*name)
	return &u
}
