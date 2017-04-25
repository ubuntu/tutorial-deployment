package consts

const (
	// GdocPrefix appending to detect google docs in refs
	GdocPrefix = "gdoc:"
	// TemplateFileName relative to metadata path
	TemplateFileName = "ubuntu-template.html"

	// APIURL is where API files are served on the webserver
	APIURL = "/api/"
	// ImagesURL is where images and other generated assets from API are served on the webserver
	ImagesURL = "/images/assets/"
	// CodelabSrcURL is where codelab source files are served on the webserver
	CodelabSrcURL = "/src/codelabs/"
	// ServeRootURL will always serve the / directory and server-side routing will do the redirect
	ServeRootURL = "/tutorial/"
)
