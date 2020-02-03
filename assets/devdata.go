// +build dev

package assets

import "net/http"

var Assets http.FileSystem = http.Dir("site")
