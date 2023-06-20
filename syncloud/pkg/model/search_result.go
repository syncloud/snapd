package model

type SearchResult struct {
	//Revision storeSearchChannelSnap `json:"revision"`
	Snap   Snap   `json:"snap"`
	Name   string `json:"name"`
	SnapID string `json:"snap-id"`
}
