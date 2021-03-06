package gameserver

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type PlatformInfoType uint8

const (
	PLATFORM_INFO_TYPE_BATTLE PlatformInfoType = 0
	PLATFORM_INFO_TYPE_BRIDGE PlatformInfoType = 1
)

type PlatformInfo struct {
	// Данные, читаемые из json
	SymbolName    string               `json:"symbol_name"`        // имя символа
	Width         uint16               `json:"width"`              // ширина
	Height        uint16               `json:"height"`             // высота
	Cells         []PlatformCellType   `json:"cells,omitempty"`    // ячейки в платформе, размер - ширина X высота
	Exits         [4]PlatformDir       `json:"exits"`              // выходы из платформы
	MonstersNames []string             `json:"monsters"`           // имена монстров на платформе
	SpawnMin      uint8                `json:"monsters_spawn_min"` // минимальное количество монстров
	SpawnMax      uint8                `json:"monsters_spawn_max"` // максимальное количество монстров
	Objects       []PlatformObjectInfo `json:"objects,omitempty"`  // объекты на платформе
	Blocks        []PlatformObjectInfo `json:"blocks,omitempty"`   // блоки?? // TODO: ???
	// Локальные данные для удобства и скорости (обрабатываются после загрузки)
	Type          PlatformInfoType                               `json:"-"` // тип платформы
	ObjectsByType map[PlatformObjectType]([]*PlatformObjectInfo) `json:"-"` // объекты по типам данных
}

// TODO: Указатели???

func NewPlatformsFromData(data []byte) (map[string]*PlatformInfo, error) {
	result := make(map[string]*PlatformInfo)
	err := json.Unmarshal(data, &result)
	if err == nil {
		for _, info := range result {
			info.handleLoadedInfo()
		}
	}
	return result, err
}

func NewPlatformsFromReader(reader io.Reader) (map[string]*PlatformInfo, error) {
	result := make(map[string]*PlatformInfo)
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&result)
	if err == nil {
		for _, info := range result {
			info.handleLoadedInfo()
		}
	}
	return result, err
}

func NewPlatformsFromFile(filePath string) (map[string]*PlatformInfo, error) {
	// Загрузка платформ из файла
	f, err := os.Open(filePath)
	if err != nil {
		log.Println(err)
		return make(map[string]*PlatformInfo), err
	}
	defer f.Close()

	return NewPlatformsFromReader(f)
}

func (info *PlatformInfo) handleLoadedInfo() {
	// Сразу определим тип, чтобы в рантайме не дергаться
	if (info.SpawnMax > 0) && (len(info.MonstersNames) > 0) {
		info.Type = PLATFORM_INFO_TYPE_BATTLE
	} else {
		info.Type = PLATFORM_INFO_TYPE_BRIDGE
	}

	// Сформируем список объектов по типам для быстрого доступа
	info.ObjectsByType = make(map[PlatformObjectType]([]*PlatformObjectInfo))
	for i := range info.Objects {
		objPtr := &(info.Objects[i])
		info.ObjectsByType[objPtr.Type] = append(info.ObjectsByType[objPtr.Type], objPtr)
	}

	// Формируем массивы ячеек для каждого объекта и блока
	for i := range info.Objects {
		obj := &(info.Objects[i])

		oldCells := obj.Cells
		obj.Cells = make([]PlatformCellType, obj.Width*obj.Height)
		for i := int16(0); i < obj.Width*obj.Height; i++ {
			if i < int16(len(oldCells)) {
				obj.Cells[i] = oldCells[i]
			} else {
				obj.Cells[i] = CELL_TYPE_SPACE
			}
		}
	}
	for i := range info.Blocks {
		obj := &(info.Blocks[i])

		oldCells := obj.Cells
		obj.Cells = make([]PlatformCellType, obj.Width*obj.Height)
		for i := int16(0); i < obj.Width*obj.Height; i++ {
			if i < int16(len(oldCells)) {
				obj.Cells[i] = oldCells[i]
			} else {
				obj.Cells[i] = CELL_TYPE_SPACE
			}
		}
	}
}
