package assets

import _ "embed"

//go:embed words.json
var WordsJSON []byte

//go:embed exists.json
var ExistsJSON []byte
