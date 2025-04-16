package prismaid

import (
	"net/http"

	"github.com/open-and-sustainable/prismaid/convert/file"
	"github.com/open-and-sustainable/prismaid/download/list"
	"github.com/open-and-sustainable/prismaid/download/zotero"
	"github.com/open-and-sustainable/prismaid/review/logic"
)

func Review(tomlConfiguration string) error {
	return logic.Review(tomlConfiguration)
}

func DownloadZoteroPDFs(username, apiKey, collectionName, parentDir string) error {
	client := &http.Client{}
	return zotero.DownloadZoteroPDFs(client, username, apiKey, collectionName, parentDir)
}

func DownloadURLList(path string) {
	list.DownloadURLList(path)
	return
}

func Convert(inputDir, selectedFormats string) error {
	return file.Convert(inputDir, selectedFormats)
}
