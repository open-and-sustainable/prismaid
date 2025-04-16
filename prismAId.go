package prismaid

import (
	"github.com/open-and-sustainable/prismaid/download/list"
	"github.com/open-and-sustainable/prismaid/download/zotero"
	"github.com/open-and-sustainable/prismaid/review/logic"
)

func RunReview(tomlConfiguration string) error {
	return logic.RunReview(tomlConfiguration)
}

func DownloadPDFs(client zotero.HttpClient, username, apiKey, collectionName, parentDir string) error {
	return zotero.DownloadPDFs(client, username, apiKey, collectionName, parentDir)
}

func RunListDownload(path string) {
	list.RunListDownload(path)
	return
}
