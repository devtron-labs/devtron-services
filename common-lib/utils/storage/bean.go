package storage

type VeleoroBslStatusUpdate struct {
	ClusterId int    `json:"clusterId"`
	BslName   string `json:"bslName"`
	Status    string `json:"status"`
}
