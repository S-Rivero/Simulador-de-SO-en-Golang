package apisSend

import (
	"github.com/sisoputnfrba/tp-golang/memoria/funcs"
	"github.com/sisoputnfrba/tp-golang/utils/commons"
	"github.com/sisoputnfrba/tp-golang/utils/globalvar"
)

func EnviarDumpAFileSystem(pid int, tid int) string {

	paquete := globalvar.DumpToFS_Type{
		Pid:   pid,
		Tid:   tid,
		Bytes: funcs.ContenidoEnRAM(pid),
	}
	if len(paquete.Bytes) == 0 {
		return "Error: No se logro obtener el contenido de memoria"
	}
	paquete.Size = len(paquete.Bytes)
	res := commons.EnviarPaqueteYEsperarOK[globalvar.DumpToFS_Type](paquete, "FS", "DUMP_MEMORY")
	if res {
		return ""
	} else {
		return "Error: FileSystema devolvio error"
	}
}
