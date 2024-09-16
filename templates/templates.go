package templates

import (
   "embed" 
)

//go:embed *.html
var EHTML embed.FS
