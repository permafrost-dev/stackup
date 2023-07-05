package lib

type Container struct {
	ID     string                 `json:"Id"`
	Exited bool                   `json:"Exited"`
	Labels map[string]interface{} `json:"Labels"`
}
