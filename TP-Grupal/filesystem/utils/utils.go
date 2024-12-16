package utils

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sisoputnfrba/tp-golang/filesystem/globals"
	"github.com/sisoputnfrba/tp-golang/utils/commons"
	"github.com/sisoputnfrba/tp-golang/utils/globalvar"
)

type FileMetadata struct {
	IndexBlock int `json:"index_block"`
	Size       int `json:"size"`
}

func createMetadataFile(filename string, metadata FileMetadata) error {
	metadataPath := filepath.Join(globals.Config.MountDir, "files", filename)

	file, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("error al identar el archivo metadata: %v", err)
	}

	return os.WriteFile(metadataPath, file, 0644)
}

func findFreeBlocks(numBlocks int) ([]int, error) {
	bitmapPath := filepath.Join(globals.Config.MountDir, "bitmap.dat")

	bitmap, err := os.ReadFile(bitmapPath)
	if err != nil {
		return nil, fmt.Errorf("error al leer el bitmap: %v", err)
	}

	freeBlocks := make([]int, 0, numBlocks)
	for byteIndex := 0; byteIndex < len(bitmap) && len(freeBlocks) < numBlocks; byteIndex++ {
		for bitIndex := 7; bitIndex >= 0 && len(freeBlocks) < numBlocks; bitIndex-- {
			// Modificamos para chequear los bits de izquierda a derecha
			if bitmap[byteIndex]&(1<<bitIndex) == 0 {
				blockNum := byteIndex*8 + (7 - bitIndex)
				freeBlocks = append(freeBlocks, blockNum)
			}
		}
	}

	log.Printf("## Bloques Libres encontrados: %d", len(freeBlocks))
	log.Printf("## Bloques Libres: %v", freeBlocks)

	if len(freeBlocks) < numBlocks {
		return nil, fmt.Errorf("no hay suficientes bloques libres: necesarios %d, encontrados %d", numBlocks, len(freeBlocks))
	}

	return freeBlocks, nil
}

func markBlocksAsUsed(blocks []int) error {
	bitmapPath := filepath.Join(globals.Config.MountDir, "bitmap.dat")

	bitmap, err := os.ReadFile(bitmapPath)
	if err != nil {
		return fmt.Errorf("error al leer el bitmap: %v", err)
	}

	log.Printf("## Bitmap antes de marcar bloques: %08b", bitmap)

	for _, block := range blocks {
		byteIndex := block / 8
		bitIndex := 7 - (block % 8)
		bitmap[byteIndex] |= 1 << bitIndex
	}

	log.Printf("## Bitmap después de marcar bloques: %08b", bitmap)

	if err := os.WriteFile(bitmapPath, bitmap, 0644); err != nil {
		return fmt.Errorf("error al escribir el bitmap: %v", err)
	}

	return nil
}

func writeBlocks(content []byte, blocks []int) error {
	bloquesPath := filepath.Join(globals.Config.MountDir, "bloques.dat")

	file, err := os.OpenFile(bloquesPath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error al abrir bloques.dat: %v", err)
	}
	defer file.Close()

	// Empezar desde el segundo bloque (el primero es el índice)
	for i := 0; i < len(blocks)-1; i++ {
		blockNum := blocks[i+1] // Usar el siguiente bloque después del índice
		offset := int64(blockNum) * int64(globals.Config.BlockSize)

		if _, err := file.Seek(offset, 0); err != nil {
			return fmt.Errorf("error al buscar bloque: %v", err)
		}

		// Calcular qué parte del contenido va en este bloque
		start := i * globals.Config.BlockSize
		end := start + globals.Config.BlockSize
		if end > len(content) {
			end = len(content)
		}

		// Si ya no hay más contenido para escribir, rellenar con ceros
		var blockContent []byte
		if start < len(content) {
			blockContent = content[start:end]
			// Si el bloque no está completo, rellenar con ceros
			if len(blockContent) < globals.Config.BlockSize {
				padding := make([]byte, globals.Config.BlockSize-len(blockContent))
				blockContent = append(blockContent, padding...)
			}
		} else {
			blockContent = make([]byte, globals.Config.BlockSize)
		}

		if _, err := file.Write(blockContent); err != nil {
			return fmt.Errorf("error al escribir bloque: %v", err)
		}

		log.Printf("## Acceso Bloque - Archivo: %d-%d.dmp - Tipo Bloque: DATOS - Bloque File System %d",
			blocks[0], blocks[1], blockNum)

		// Aplicar el retardo configurado
		time.Sleep(time.Duration(globals.Config.BlockAccessDelay) * time.Millisecond)
	}

	return nil
}

func writeIndexBlock(indexBlock int, dataBlocks []int) error {
	bloquesPath := filepath.Join(globals.Config.MountDir, "bloques.dat")

	file, err := os.OpenFile(bloquesPath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error al intentar abrir bloques.dat: %v", err)
	}
	defer file.Close()

	offset := int64(indexBlock) * int64(globals.Config.BlockSize)
	if _, err := file.Seek(offset, 0); err != nil {
		return fmt.Errorf("error al buscar indice de bloque: %v", err)
	}

	// Escribir los punteros a los bloques de datos
	for _, block := range dataBlocks[1:] {
		if err := binary.Write(file, binary.LittleEndian, uint32(block)); err != nil {
			return fmt.Errorf("error al escribir puntero de bloque: %v", err)
		}
	}

	time.Sleep(time.Duration(globals.Config.BlockAccessDelay) * time.Millisecond)
	log.Printf("## Acceso Bloque - Archivo: index - Tipo Bloque: INDICE - Bloque File System %d", indexBlock)

	return nil
}

func HandleCreateDump(w http.ResponseWriter, r *http.Request) {
	var req globalvar.DumpToFS_Type
	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}

	// Log total blocks in the filesystem
	log.Printf("## Bloques totales de filesystem: %d", globals.Config.BlockCount)

	// Calculate the number of blocks needed
	numDataBlocks := int(math.Ceil(float64(req.Size) / float64(globals.Config.BlockSize)))
	totalBlocks := numDataBlocks + 1 // +1 for the index block

	// Log the file size and required blocks
	log.Printf("## Tamaño Archivo: %d - Bloques Necesarios: %d", req.Size, totalBlocks)

	// Find free blocks
	blocks, err := findFreeBlocks(totalBlocks)

	// Log the number of free blocks found
	log.Printf("## Bloques Necesarios: %d - Bloques Libres encontrados: %d", totalBlocks, len(blocks))
	log.Printf("## Bloques Libres: %v", blocks)

	if err != nil {
		log.Printf("No hay suficiente espacio disponible: %v", err)
		http.Error(w, "No hay espacio disponible", http.StatusInsufficientStorage)
		return
	}

	// Check if there are enough free blocks
	if len(blocks) < totalBlocks {
		log.Printf("No hay suficientes bloques libres: Necesarios: %d, Libres: %d", totalBlocks, len(blocks))
		http.Error(w, "No hay espacio disponible", http.StatusInsufficientStorage)
		return
	}

	// Create filename with timestamp
	timestamp := time.Now().Format("15:04:05:000")
	filename := fmt.Sprintf("%d-%d-%s.dmp", req.Pid, req.Tid, timestamp)

	// Create metadata
	metadata := FileMetadata{
		IndexBlock: blocks[0],
		Size:       req.Size,
	}

	// Create metadata file
	if err := createMetadataFile(filename, metadata); err != nil {
		log.Printf("Error al crear archivo de metadata: %v", err)
		http.Error(w, "Error al crear archivo de metadata", http.StatusInternalServerError)
		return
	}

	// Mark blocks as used
	if err := markBlocksAsUsed(blocks); err != nil {
		log.Printf("Error al marcar bloques como usados: %v", err)
		http.Error(w, "Error al marcar bloques como usados", http.StatusInternalServerError)
		return
	}

	// Write the index block
	if err := writeIndexBlock(blocks[0], blocks); err != nil {
		log.Printf("Error al escribir el bloque índice: %v", err)
		http.Error(w, "Error al escribir el bloque índice", http.StatusInternalServerError)
		return
	}

	// Write the data blocks
	if err := writeBlocks(req.Bytes, blocks); err != nil {
		log.Printf("Error al escribir los bloques de datos: %v", err)
		http.Error(w, "Error al escribir los bloques de datos", http.StatusInternalServerError)
		return
	}

	// Log the created file and its size
	log.Printf("## Archivo Creado: %s - Tamaño: %d", filename, req.Size)

	// Calculate and log the number of remaining free blocks
	remainingBlocks, err := countFreeBlocks()
	if err != nil {
		log.Printf("Error al contar los bloques libres: %v", err)
		http.Error(w, "Error al contar los bloques libres", http.StatusInternalServerError)
		return
	}
	log.Printf("## Bloques Libres después de asignar: %d", remainingBlocks)
	for i, block := range blocks {
		if i == 0 {
			log.Printf("## Bloque asignado: %d - Archivo: %s - Bloques Libres: %d", block, filename, remainingBlocks)
		}
	}
	log.Printf("## Fin de solicitud - Archivo: %s", filename)

	w.WriteHeader(http.StatusOK)
	//json.NewEncoder(w).Encode(map[string]string{"filename": filename})
}

func countFreeBlocks() (int, error) {
	bitmapPath := filepath.Join(globals.Config.MountDir, "bitmap.dat")

	bitmap, err := os.ReadFile(bitmapPath)
	if err != nil {
		return 0, fmt.Errorf("error al leer el bitmap: %v", err)
	}

	freeBlocks := 0
	for byteIndex := 0; byteIndex < len(bitmap); byteIndex++ {
		for bitIndex := 7; bitIndex >= 0; bitIndex-- {
			if bitmap[byteIndex]&(1<<bitIndex) == 0 {
				freeBlocks++
			}
		}
	}

	return freeBlocks, nil
}

func FSInit() error {
	log.Printf("Iniciando sistema de archivos en: %s", globals.Config.MountDir)

	pathRemove := filepath.Join(globals.Config.MountDir, "bloques.dat")
	os.Remove(pathRemove)
	pathRemove = filepath.Join(globals.Config.MountDir, "bitmap.dat")
	os.Remove(pathRemove)

	// Verificar/crear directorio de montaje
	if err := createMountDir(); err != nil {
		return fmt.Errorf("error creando directorio de montaje: %v", err)
	}

	// Crear subdirectorio /files si no existe
	filesDir := filepath.Join(globals.Config.MountDir, "files")
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de files: %v", err)
	}

	// Verificar/crear bitmap.dat
	if err := initBitmap(); err != nil {
		return fmt.Errorf("error inicializando bitmap: %v", err)
	}

	// Verificar/crear bloques.dat
	if err := initBloques(); err != nil {
		return fmt.Errorf("error inicializando bloques: %v", err)
	}

	log.Printf("Sistema de archivos inicializado correctamente en %s", globals.Config.MountDir)
	return nil
}

// createMountDir crea el directorio de montaje si no existe
func createMountDir() error {
	if err := os.MkdirAll(globals.Config.MountDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de montaje: %v", err)
	}
	log.Printf("Directorio de montaje verificado: %s", globals.Config.MountDir)
	return nil
}

func initBitmap() error {
	bitmapPath := filepath.Join(globals.Config.MountDir, "bitmap.dat")

	// Calcular tamaño del bitmap (redondeado hacia arriba)
	bitmapSize := int(math.Ceil(float64(globals.Config.BlockCount) / 8.0))

	// Verificar si el archivo existe
	if _, err := os.Stat(bitmapPath); os.IsNotExist(err) {
		// Crear nuevo archivo bitmap
		file, err := os.Create(bitmapPath)
		if err != nil {
			return fmt.Errorf("error creando bitmap.dat: %v", err)
		}
		defer file.Close()

		// Inicializar con ceros (bloques libres)
		zeroBytes := make([]byte, bitmapSize)
		if _, err := file.Write(zeroBytes); err != nil {
			return fmt.Errorf("error inicializando bitmap.dat: %v", err)
		}

		log.Printf("bitmap.dat creado e inicializado con tamaño %d bytes", bitmapSize)
	} else {
		// Verificar tamaño del archivo existente
		fileInfo, err := os.Stat(bitmapPath)
		if err != nil {
			return fmt.Errorf("error verificando bitmap.dat: %v", err)
		}
		if fileInfo.Size() != int64(bitmapSize) {
			return fmt.Errorf("tamaño incorrecto de bitmap.dat: esperado %d, actual %d", bitmapSize, fileInfo.Size())
		}
		log.Printf("bitmap.dat existente verificado: %d bytes", fileInfo.Size())
	}
	return nil
}

func initBloques() error {
	bloquesPath := filepath.Join(globals.Config.MountDir, "bloques.dat")

	// Calcular tamaño total del archivo de bloques
	bloquesSize := globals.Config.BlockCount * globals.Config.BlockSize

	// Verificar si el archivo existe
	if _, err := os.Stat(bloquesPath); os.IsNotExist(err) {
		// Crear nuevo archivo de bloques
		file, err := os.Create(bloquesPath)
		if err != nil {
			return fmt.Errorf("error creando bloques.dat: %v", err)
		}
		defer file.Close()

		// Reservar espacio para todos los bloques
		zeroBytes := make([]byte, bloquesSize)
		if _, err := file.Write(zeroBytes); err != nil {
			return fmt.Errorf("error inicializando bloques.dat: %v", err)
		}

		log.Printf("bloques.dat creado e inicializado con tamaño %d bytes", bloquesSize)
	} else {
		// Verificar tamaño del archivo existente
		fileInfo, err := os.Stat(bloquesPath)
		if err != nil {
			return fmt.Errorf("error verificando bloques.dat: %v", err)
		}
		if fileInfo.Size() != int64(bloquesSize) {
			return fmt.Errorf("tamaño incorrecto de bloques.dat: esperado %d, actual %d", bloquesSize, fileInfo.Size())
		}
		log.Printf("bloques.dat existente verificado: %d bytes", fileInfo.Size())
	}
	return nil
}
