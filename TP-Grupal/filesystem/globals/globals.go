package globals

type ModuleConfig struct {
	Ip               string `json:"ip"`
	IpMemory         string `json:"ip_memory"`
	PortMemory       int    `json:"port_memory"`
	MountDir         string `json:"mount_dir"`
	BlockSize        int    `json:"block_size"`
	BlockCount       int    `json:"block_count"`
	BlockAccessDelay int    `json:"block_access_delay"`
	Port             int    `json:"port"`
	Name             string `json:"name"`
}

var Config *ModuleConfig

type FileInfo struct {
	IndexBlock int `json:"index_block"` // Número de bloque que corresponde al bloque de índices
	Size       int `json:"size"`        // Tamaño del archivo en bytes
}
