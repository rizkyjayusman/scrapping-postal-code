package clientbps

type Client interface {
	GetRegion(parent string, level int) ([]ResponseBodyGetRegion, error)
}

type ResponseBodyGetRegion struct {
	KodeBps string `json:"kode_bps"`
	NamaBps string `json:"nama_bps"`
	KodePos string `json:"kode_pos"`
	NamaPos string `json:"nama_pos"`
}
