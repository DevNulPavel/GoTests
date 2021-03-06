package gameserver

import (
	"log"
	"math/rand"
)

const (
	PLATFORM_SIDE_SIZE      = 24 // вся платформа с учетом мостов
	PLATFORM_WORK_SIZE      = 18 // платформа без мостов
	PLATFORM_BLOCK_SIZE_3x3 = 3  // Platform Block Size 3х3
	PLATFORM_BLOCK_SIZE_6x6 = 6  // Platform Block Size 6х6
)

type PlatformMonster struct {
	Name string `json:"name"`
	X    int16  `json:"x"`
	Y    int16  `json:"y"`
}

type PlatformObject struct {
	Id  string  `json:"id"`
	X   float64 `json:"x"`
	Y   float64 `json:"y"`
	Rot int8    `json:"r"`
}

type Platform struct {
	Info *PlatformInfo `json:"-"`
	// Pos and size
	PosX   int16  `json:"x"`
	PosY   int16  `json:"y"`
	Width  uint16 `json:"width"`
	Height uint16 `json:"height"`
	// Enter
	EnterX   int16       `json:"enterX"`
	EnterY   int16       `json:"enterY"`
	EnterDir PlatformDir `json:"enterDir"`
	// Exit
	ExitCoord [4]int16    `json:"exit"`
	ExitDir   PlatformDir `json:"exitDir"`
	// Symbol name
	SymbolName string `json:"symbolName"`
	// Bridge
	IsBridge bool `json:"isBridge"`
	// Monsters
	MonsterSpawnMin  uint8    `json:"monsterSpawnMin"`    // TODO: ???
	MonsterSpawnMax  uint8    `json:"monsterSpawnMax"`    // TODO: ???
	PossibleMonsters []string `json:"monsters,omitempty"` // TODO: ???
	// Cells
	Cells []PlatformCellType `json:"cells,omitempty"`
	// Items
	Objects   []PlatformObject `json:"objects,omitempty"` // TODO: ???
	Blocks    []PlatformObject `json:"blocks,omitempty"`  // TODO: ???
	HaveDecor bool             `json:"withDecor"`
}

func NewPlatform(info *PlatformInfo, posX, posY int16, exits [4]int16, isBridge bool) *Platform {
	platform := &Platform{}

	// Info
	platform.Info = info

	// Pos and size
	platform.PosX = posX
	platform.PosY = posY
	platform.Width = info.Width
	platform.Height = info.Height

	// Exit and enter
    platform.ExitCoord = exits
	for i, coord := range platform.ExitCoord {
		if coord != -1 {
			dir := PlatformDir(i)

			point := getPortalCoord(dir, exits)

			platform.EnterX = posX + point.X
			platform.EnterY = posY + point.Y
			platform.EnterDir = dir
			break
		}
	}

	// Symbol name
	platform.SymbolName = info.SymbolName

	// Bridge
	platform.IsBridge = info.Type == PLATFORM_INFO_TYPE_BRIDGE

	// Monster count
	platform.MonsterSpawnMin = info.SpawnMin
	platform.MonsterSpawnMax = info.SpawnMax

    // Monster names list
    //platform.PossibleMonsters = make([]string, len(info.MonstersNames))
    //copy(platform.PossibleMonsters, info.MonstersNames)
	platform.PossibleMonsters = append(platform.PossibleMonsters, info.MonstersNames...)

	// Cells and walls
	createCells(platform, isBridge)

	return platform
}

func getPortalCoord(dir PlatformDir, exit [4]int16) Point16 {
	switch {
	case (dir == DIR_NORTH) && (exit[DIR_NORTH] != -1):
		return Point16{exit[DIR_NORTH], 0}

	case (dir == DIR_EAST) && (exit[DIR_EAST] != -1):
		return Point16{PLATFORM_SIDE_SIZE - 1, exit[DIR_EAST]}

	case (dir == DIR_SOUTH) && (exit[DIR_SOUTH] != -1):
		return Point16{exit[DIR_SOUTH], PLATFORM_SIDE_SIZE - 1}

	case (dir == DIR_WEST) && (exit[DIR_WEST] != -1):
		return Point16{0, exit[DIR_WEST]}

	default:
		return Point16{-1, -1}
	}

	return Point16{-1, -1}
}

func createCells(platform *Platform, isBridge bool) {
	// TODO: разделить??
	if isBridge {
		makeBridgeCells(platform)
	} else {
		makeBattleCells(platform)
	}
	//makeTestCells(platform)
}

func makeTestCells(platform *Platform) {
	platform.EnterX = 10
	platform.EnterY = 2
	platform.EnterDir = DIR_EAST

	// Make cells
	platform.Cells = make([]PlatformCellType, PLATFORM_SIDE_SIZE*PLATFORM_SIDE_SIZE)
	for i := range platform.Cells {
		x := i % int(platform.Width)
		//y := i % int(platform.Width)

		if int16(x) == platform.EnterX {
			platform.Cells[i] = CELL_TYPE_SPACE
		} else {
			platform.Cells[i] = CELL_TYPE_BLOCK
		}
	}
}

func makeBridgeCells(platform *Platform) {
	w := platform.Width
	h := platform.Height

	// Blocks
	block3x3 := make([]*PlatformObjectInfo, 0)
	for i := range platform.Info.Blocks {
		obj := &platform.Info.Blocks[i]
		if (obj.Width == PLATFORM_BLOCK_SIZE_3x3) && (obj.Height == PLATFORM_BLOCK_SIZE_3x3) {
			block3x3 = append(block3x3, obj)
		}
	}

	// Info
	cellsInfo := [PLATFORM_SIDE_SIZE * PLATFORM_SIDE_SIZE]PlatformCellType{}
	for i := 0; i < PLATFORM_SIDE_SIZE*PLATFORM_SIDE_SIZE; i++ {
		cellsInfo[i] = CELL_TYPE_BLOCK
	}

	for y := uint16(0); y < h; y += PLATFORM_BLOCK_SIZE_3x3 {
		for x := uint16(0); x < w; x += PLATFORM_BLOCK_SIZE_3x3 {
			haveBlock := false
			for yy := uint16(0); yy < PLATFORM_BLOCK_SIZE_3x3; yy++ {
				for xx := uint16(0); xx < PLATFORM_BLOCK_SIZE_3x3; xx++ {
					// Index
					cellsIndex := (y+yy)*w + (x + xx)
					infoCellValue := platform.Info.Cells[cellsIndex]
					// Update cells
					cellsInfo[cellsIndex] = infoCellValue
					if infoCellValue != CELL_TYPE_BLOCK {
						haveBlock = true
					}
				}
			}

			if haveBlock {
				platform.Blocks, _ = appendObjects(platform.Blocks, block3x3,
					float64(x), float64(y),
					int8((x+y)&3), 3)
			}
		}
	}

	// Make cells
	cellsCount := w * h
	platform.Cells = make([]PlatformCellType, cellsCount)
	for i := range platform.Cells {
		platform.Cells[i] = CELL_TYPE_BLOCK
	}
	for i := uint16(0); i < cellsCount; i++ {
		platform.Cells[i] = cellsInfo[i]
	}
}

func makeBattleCells(platform *Platform) {
	endPoints := make([]Point16, 0)

	w := platform.Width
	h := platform.Height

	// Blocks
	block3x3 := make([]*PlatformObjectInfo, 0)
	block6x6 := make([]*PlatformObjectInfo, 0)
	for i := range platform.Info.Blocks {
		obj := &(platform.Info.Blocks[i])
		if (obj.Width == PLATFORM_BLOCK_SIZE_3x3) && (obj.Height == PLATFORM_BLOCK_SIZE_3x3) {
			block3x3 = append(block3x3, obj)
		} else if (obj.Width == PLATFORM_BLOCK_SIZE_6x6) && (obj.Height == PLATFORM_BLOCK_SIZE_6x6) {
			block6x6 = append(block6x6, obj)
		}
	}

	// Info
    cellsCount := w*h
	cellsInfo := make([]PlatformCellType, cellsCount)
	cellsWalls := make([]PlatformCellType, cellsCount)
	for i := uint16(0); i < cellsCount; i++ {
		cellsInfo[i] = CELL_TYPE_BLOCK
		cellsWalls[i] = CELL_TYPE_UNDEF
	}

	// Make cells
	platform.Cells = make([]PlatformCellType, PLATFORM_SIDE_SIZE*PLATFORM_SIDE_SIZE)
	for i := range platform.Cells {
		platform.Cells[i] = CELL_TYPE_BLOCK
	}

	// Make logic
	foundPath := false
	for foundPath == false {
		// Clear arrays
		platform.Blocks = make([]PlatformObject, 0)
		platform.Objects = make([]PlatformObject, 0)

		// Clear info
		for i := 0; i < PLATFORM_SIDE_SIZE*PLATFORM_SIDE_SIZE; i++ {
			cellsInfo[i] = CELL_TYPE_BLOCK
			cellsWalls[i] = CELL_TYPE_UNDEF
		}

		createBlocks6x6(platform, cellsInfo, cellsWalls, block6x6)
		createBlocks3x3(platform, cellsInfo, block3x3)

		createArches(platform, cellsInfo, cellsWalls)
		createWalls(platform, cellsInfo, cellsWalls)

		// заполняем стенами ячейки
		for y := uint16(0); y < PLATFORM_WORK_SIZE; y++ {
			for x := uint16(0); x < PLATFORM_WORK_SIZE; x++ {
				index := uint16(y*w + x)
				if cellsWalls[index] != CELL_TYPE_UNDEF {
					cellsInfo[index] = cellsWalls[index]
				}
			}
		}

		createPlatformElements(platform, cellsInfo)

		start := -1
		foundPath = false

		// TODO: ??? было с 1цы
		for i := 0; i < 4; i++ {
			if platform.ExitCoord[i] != -1 {
				if start == -1 {
					start = i
				} else {
					endPoints = make([]Point16, 0, 1)

					endPoint := getPortalCoord(PlatformDir(start), platform.ExitCoord)
					endPoints = append(endPoints, endPoint)

					// TODO: ???
					/*
						startPoint := getPortalCoord(PlatformDir(start), platform.ExitCoord)

						path = _pathManager.findPathOld(startpt, endPoints, false, true, false);
						if len(path) > 0 {
							foundPath = true
							break
						}*/
					foundPath = true
					break
				}
			}
		}
	}

	for i := 0; i < 4; i++ {
		dir := PlatformDir(i)

		exitPoint := getPortalCoord(dir, platform.ExitCoord)
		if (exitPoint.X == -1) || (exitPoint.Y == -1) {
			continue
		}

		y := exitPoint.Y
		x := exitPoint.X

		if dir == DIR_EAST {
			x -= PLATFORM_BLOCK_SIZE_6x6
		}
		if dir == DIR_SOUTH {
			y -= PLATFORM_BLOCK_SIZE_6x6
		}

		if (dir == DIR_NORTH) || (dir == DIR_SOUTH) {
			cellsInfo[y*int16(w)+(x-2)] = CELL_TYPE_WALL
			cellsInfo[y*int16(w)+(x-1)] = CELL_TYPE_WALL
			cellsInfo[y*int16(w)+(x+1)] = CELL_TYPE_WALL
			cellsInfo[y*int16(w)+(x+2)] = CELL_TYPE_WALL
		}
		if (dir == DIR_EAST) || (dir == DIR_WEST) {
			cellsInfo[(y-2)*int16(w)+x] = CELL_TYPE_WALL
			cellsInfo[(y-1)*int16(w)+x] = CELL_TYPE_WALL
			cellsInfo[(y+1)*int16(w)+x] = CELL_TYPE_WALL
			cellsInfo[(y+2)*int16(w)+x] = CELL_TYPE_WALL
		}
	}

	for i := uint16(0); i < w*h; i++ {
		platform.Cells[i] = cellsInfo[i]
	}
}

func createBlocks6x6(platform *Platform, cellInfo, cellWalls []PlatformCellType, block6x6 []*PlatformObjectInfo) {
	platform.HaveDecor = false

	for y := int16(0); y < PLATFORM_WORK_SIZE; y += PLATFORM_BLOCK_SIZE_6x6 {
		for x := int16(0); x < PLATFORM_WORK_SIZE; x += PLATFORM_BLOCK_SIZE_6x6 {
			curPoint := NewPoint16(x, y)
			curPoint = curPoint.Div(PLATFORM_BLOCK_SIZE_6x6)

			northPoint := getPortalCoord(DIR_NORTH, platform.ExitCoord)
			eastPoint := getPortalCoord(DIR_EAST, platform.ExitCoord)
			southPoint := getPortalCoord(DIR_SOUTH, platform.ExitCoord)
			westPoint := getPortalCoord(DIR_WEST, platform.ExitCoord)

			testPoint1 := northPoint.Div(PLATFORM_BLOCK_SIZE_6x6)

			testPoint2 := NewPoint16(eastPoint.X-PLATFORM_BLOCK_SIZE_6x6, eastPoint.Y)
			testPoint2 = testPoint2.Div(PLATFORM_BLOCK_SIZE_6x6)

			testPoint3 := NewPoint16(southPoint.X, southPoint.Y-PLATFORM_BLOCK_SIZE_6x6)
			testPoint3 = testPoint3.Div(PLATFORM_BLOCK_SIZE_6x6)

			testPoint4 := westPoint.Div(PLATFORM_BLOCK_SIZE_6x6)

			isExit := false
			if (curPoint == testPoint1) || (curPoint == testPoint2) || (curPoint == testPoint3) || (curPoint == testPoint4) {
				isExit = true
			}

			posTest := (y == PLATFORM_WORK_SIZE/2-PLATFORM_BLOCK_SIZE_3x3) && (x == PLATFORM_WORK_SIZE/2-PLATFORM_BLOCK_SIZE_3x3)
			if (rand.Int()%2 == 0) || posTest || ((rand.Int()%2 == 0) && isExit) {
				// TODO: править тут
				newArray, item := appendObjects(platform.Blocks,
					block6x6,
					float64(x), float64(y),
					int8((x+y)&3), 3)
				platform.Blocks = newArray

				for yy := int16(0); yy < item.Height; yy++ {
					for xx := int16(0); xx < item.Width; xx++ {
						cellInfo[(y+yy)*PLATFORM_WORK_SIZE+(x+xx)] = CELL_TYPE_SPACE
						if item.Cells[yy*item.Width+xx] == CELL_TYPE_BLOCK {
							cellWalls[(y+yy)*PLATFORM_WORK_SIZE+(x+xx)] = CELL_TYPE_PIT
						}
					}
				}

				// если можем, то применяем декор
				if ((y == PLATFORM_WORK_SIZE/2-PLATFORM_BLOCK_SIZE_3x3) &&
					(x == PLATFORM_WORK_SIZE/2-PLATFORM_BLOCK_SIZE_3x3)) &&
					(rand.Int()%3 == 0) {

					platform.Objects, _ = appendObjects(platform.Objects,
						platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_DECOR],
						float64(x), float64(y),
						0, 3)

					platform.HaveDecor = true
				}
			}
		}
	}
}

// TODO: Пробрасывается ли указатель в cellInfo??
func createBlocks3x3(platform *Platform, cellInfo []PlatformCellType, block3x3 []*PlatformObjectInfo) {
	edges := make([]Point16, 0)
	for y := int16(0); y < PLATFORM_WORK_SIZE; y += PLATFORM_BLOCK_SIZE_3x3 {
		for x := int16(0); x < PLATFORM_WORK_SIZE; x += PLATFORM_BLOCK_SIZE_3x3 {
			addEdge := false
			addEdge = addEdge || (x == 0)
			addEdge = addEdge || (y == 0)
			addEdge = addEdge || (x == PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_3x3)
			addEdge = addEdge || (y == PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_3x3)

			validCell := cellInfo[y*int16(platform.Width)+x] == CELL_TYPE_BLOCK

			if addEdge && validCell {
				edges = append(edges, NewPoint16(x, y))
			}
		}
	}
	// Перемешивание
	for i := range edges {
		j := rand.Intn(i + 1)
		edges[i], edges[j] = edges[j], edges[i]
	}

	for i := 0; i < 4; i++ {
		dir := PlatformDir(i)
		exit := getPortalCoord(dir, platform.ExitCoord)
		if (exit.X == -1) || (exit.Y == -1) {
			continue
		}
		// TODO: было в оригинале
		exit = exit.Div(3)
		exit = exit.Mul(3)

		if cellInfo[exit.Y*int16(platform.Width)+exit.X] == CELL_TYPE_BLOCK {
			for yy := int16(0); yy < PLATFORM_BLOCK_SIZE_3x3; yy++ {
				for xx := int16(0); xx < PLATFORM_BLOCK_SIZE_3x3; xx++ {
					index := (exit.Y+yy)*int16(platform.Width) + (exit.X + xx)
					cellInfo[index] = CELL_TYPE_SPACE
				}
			}
			if (rand.Int()%3 == 0) && (dir == DIR_EAST || dir == DIR_SOUTH) {
				direction := int8(1)
				if (i & 1) != 0 {
					direction = 0
				}
				platform.Blocks, _ = appendObjects(platform.Blocks,
					platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_FLOOR],
					float64(exit.X), float64(exit.Y),
					direction,
					3)
			} else {
				direction := int8((exit.X + exit.Y) & 3)
				platform.Blocks, _ = appendObjects(platform.Blocks,
					block3x3,
					float64(exit.X), float64(exit.Y),
					direction,
					3)
			}
		}
	}

	exit := getPortalCoord(DIR_EAST, platform.ExitCoord)
	if exit.X != -1 {
		edges = append(edges, NewPoint16(exit.X-2, exit.Y-1))
	}
	exit = getPortalCoord(DIR_SOUTH, platform.ExitCoord)
	if exit.X != -1 {
		edges = append(edges, NewPoint16(exit.X-1, exit.Y-2))
	}

	center := NewPoint16(PLATFORM_WORK_SIZE/2-PLATFORM_BLOCK_SIZE_3x3, PLATFORM_WORK_SIZE/2-PLATFORM_BLOCK_SIZE_3x3)

	edgesSize := len(edges)
	for i := 0; i < edgesSize; i++ {
		// Iterate from end
		point := edges[len(edges)-1]
		edges = edges[0 : len(edges)-1]

		searchComplete := false
		for searchComplete {
			// Check1
			check1 := false
			check1 = check1 || (rand.Int()%2 == 0)
			check1 = check1 || (point.Y/PLATFORM_BLOCK_SIZE_6x6 == center.Y/PLATFORM_BLOCK_SIZE_6x6)
			check1 = check1 || (point.X >= (PLATFORM_WORK_SIZE - PLATFORM_BLOCK_SIZE_3x3))
			// Check2
			check2 := point.X/PLATFORM_BLOCK_SIZE_6x6 != center.X/PLATFORM_BLOCK_SIZE_6x6
			// Check3
			check3 := point.Y < (PLATFORM_WORK_SIZE - PLATFORM_BLOCK_SIZE_3x3)
			if check1 && check2 && check3 {
				if point.X/PLATFORM_BLOCK_SIZE_6x6 < center.X/PLATFORM_BLOCK_SIZE_6x6 {
					point.X += PLATFORM_BLOCK_SIZE_3x3
				} else {
					point.X -= PLATFORM_BLOCK_SIZE_3x3
				}
			} else {
				if point.Y/PLATFORM_BLOCK_SIZE_6x6 < center.Y/PLATFORM_BLOCK_SIZE_6x6 {
					point.Y += PLATFORM_BLOCK_SIZE_3x3
				} else {
					point.Y -= PLATFORM_BLOCK_SIZE_3x3
				}
			}

			check4 := point.Div(PLATFORM_BLOCK_SIZE_6x6) != center.Div(PLATFORM_BLOCK_SIZE_6x6)
			check5 := cellInfo[point.Y*int16(platform.Width)+point.X] == CELL_TYPE_BLOCK
			if check4 && check5 {
				for yy := int16(0); yy < PLATFORM_BLOCK_SIZE_3x3; yy++ {
					for xx := int16(0); xx < PLATFORM_BLOCK_SIZE_3x3; xx++ {
						cellInfo[(exit.Y+yy)*int16(platform.Width)+(exit.X+xx)] = CELL_TYPE_SPACE
					}
				}
			}

			// TODO: ???
			if point.Div(PLATFORM_BLOCK_SIZE_6x6) == center.Div(PLATFORM_BLOCK_SIZE_6x6) {
				searchComplete = true
			}
		}
	}
}

// TODO: Пробрасывается ли указатель в cellInfo + cellsWals??
func createArches(platform *Platform, cellInfo, cellsWalls []PlatformCellType) {
	for i := 0; i < 4; i++ {
		if rand.Int()%2 == 0 {
			continue
		}

		dir := PlatformDir(i)
		exit := getPortalCoord(dir, platform.ExitCoord)
		if (exit.X == -1) || (exit.Y == -1) {
			continue
		}

		x := exit.X
		y := exit.Y

		if dir == DIR_EAST {
			x -= PLATFORM_BLOCK_SIZE_6x6
		} else if dir == DIR_SOUTH {
			y -= PLATFORM_BLOCK_SIZE_6x6
		}

		// TODO: оптимизации

		check1 := (y > 1) && (cellInfo[(y-4)*int16(platform.Width)+x] == CELL_TYPE_SPACE)
		check2 := (y != (PLATFORM_WORK_SIZE - 2)) && (cellInfo[(y+2)*int16(platform.Width)+x] == CELL_TYPE_SPACE)
		if (dir == DIR_WEST) && check1 && check2 {
			platform.Objects, _ = appendObjects(platform.Objects,
				platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_ARCHE],
				float64(x), float64(y)-2.5,
				int8(DIR_NORTH), 3)

			cellsWalls[(y-4)*int16(platform.Width)+x] = CELL_TYPE_WALL
			cellsWalls[(y-3)*int16(platform.Width)+x] = CELL_TYPE_WALL
			cellsWalls[(y-2)*int16(platform.Width)+x] = CELL_TYPE_WALL
			cellsWalls[(y+2)*int16(platform.Width)+x] = CELL_TYPE_WALL
			cellsWalls[(y+3)*int16(platform.Width)+x] = CELL_TYPE_WALL
			cellsWalls[(y+4)*int16(platform.Width)+x] = CELL_TYPE_WALL
			continue
		}

		//check1 = (y > 1) && (cellInfo[(y-4)*int16(platform.Width) + x] == CELL_TYPE_SPACE)
		//check2 = (y != (PLATFORM_WORK_SIZE-2)) && (cellInfo[(y + 2)*int16(platform.Width)+x] == CELL_TYPE_SPACE)
		if (dir == DIR_EAST) && check1 && check2 {
			platform.Objects, _ = appendObjects(platform.Objects,
				platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_ARCHE],
				float64(x)-2, float64(y)+0.5,
				int8(DIR_SOUTH), 3)

			cellsWalls[(y-4)*int16(platform.Width)+x] = CELL_TYPE_WALL
			cellsWalls[(y-3)*int16(platform.Width)+x] = CELL_TYPE_WALL
			cellsWalls[(y-2)*int16(platform.Width)+x] = CELL_TYPE_WALL
			cellsWalls[(y+2)*int16(platform.Width)+x] = CELL_TYPE_WALL
			cellsWalls[(y+3)*int16(platform.Width)+x] = CELL_TYPE_WALL
			cellsWalls[(y+4)*int16(platform.Width)+x] = CELL_TYPE_WALL
			continue
		}

		check1 = (x > 1) && (cellInfo[y*int16(platform.Width)+(x-4)] == CELL_TYPE_SPACE)
		check2 = (x != (PLATFORM_WORK_SIZE - 2)) && (cellInfo[y*int16(platform.Width)+(x+2)] == CELL_TYPE_SPACE)
		if (dir == DIR_NORTH) && check1 && check2 {
			platform.Objects, _ = appendObjects(platform.Objects,
				platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_ARCHE],
				float64(x)-2.5, float64(y)-1.5,
				int8(DIR_EAST), 3)

			cellsWalls[y*int16(platform.Width)+(x-4)] = CELL_TYPE_WALL
			cellsWalls[y*int16(platform.Width)+(x-3)] = CELL_TYPE_WALL
			cellsWalls[y*int16(platform.Width)+(x-2)] = CELL_TYPE_WALL
			cellsWalls[y*int16(platform.Width)+(x+2)] = CELL_TYPE_WALL
			cellsWalls[y*int16(platform.Width)+(x+3)] = CELL_TYPE_WALL
			cellsWalls[y*int16(platform.Width)+(x+4)] = CELL_TYPE_WALL
			continue
		}

		//check1 = (x > 1) && (cellInfo[y*int16(platform.Width) + (x-4)] == CELL_TYPE_SPACE)
		//check2 = (x != (PLATFORM_WORK_SIZE-2)) && (cellInfo[y*int16(platform.Width)+(x+2)] == CELL_TYPE_SPACE)
		if (dir == DIR_SOUTH) && check1 && check2 {
			platform.Objects, _ = appendObjects(platform.Objects,
				platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_ARCHE],
				float64(x)+0.5, float64(y)-0.5,
				int8(DIR_WEST), 3)

			cellsWalls[y*int16(platform.Width)+(x-4)] = CELL_TYPE_WALL
			cellsWalls[y*int16(platform.Width)+(x-3)] = CELL_TYPE_WALL
			cellsWalls[y*int16(platform.Width)+(x-2)] = CELL_TYPE_WALL
			cellsWalls[y*int16(platform.Width)+(x+2)] = CELL_TYPE_WALL
			cellsWalls[y*int16(platform.Width)+(x+3)] = CELL_TYPE_WALL
			cellsWalls[y*int16(platform.Width)+(x+4)] = CELL_TYPE_WALL
			continue
		}
	}
}

// TODO: Пробрасывается ли указатель в cellInfo + cellsWals??
func createWalls(platform *Platform, cellInfo, cellsWalls []PlatformCellType) {
	xMax := int16(platform.Height - PLATFORM_BLOCK_SIZE_6x6)
	yMax := int16(platform.Height - PLATFORM_BLOCK_SIZE_6x6)

	for y := int16(0); y < yMax; y += PLATFORM_BLOCK_SIZE_3x3 {
		for x := int16(0); x < xMax; x += PLATFORM_BLOCK_SIZE_3x3 {
			point := NewPoint16(x, y)

			// Выход
			{
				test1 := getPortalCoord(DIR_NORTH, platform.ExitCoord)

				north := getPortalCoord(DIR_NORTH, platform.ExitCoord)
				test2 := NewPoint16(north.X-PLATFORM_BLOCK_SIZE_6x6, north.Y)

				south := getPortalCoord(DIR_SOUTH, platform.ExitCoord)
				test3 := NewPoint16(south.X, south.Y-PLATFORM_BLOCK_SIZE_6x6)

				test4 := getPortalCoord(DIR_WEST, platform.ExitCoord)

				if (point.Distance(test1) <= 4.5) ||
					(point.Distance(test2) <= 4.5) ||
					(point.Distance(test3) <= 4.5) ||
					(point.Distance(test4) <= 4.5) {
					continue
				}
			}

			// дальше от центра
			{
				test1 := (x > PLATFORM_BLOCK_SIZE_6x6) && (x < PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_6x6) && (y > PLATFORM_BLOCK_SIZE_6x6) && (y < PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_6x6)
				test2 := (y/PLATFORM_BLOCK_SIZE_6x6 == PLATFORM_WORK_SIZE/2/PLATFORM_BLOCK_SIZE_6x6) && (x/PLATFORM_BLOCK_SIZE_6x6 == PLATFORM_WORK_SIZE/PLATFORM_BLOCK_SIZE_6x6/PLATFORM_BLOCK_SIZE_6x6)
				if test1 || test2 {
					continue
				}
			}

			// верхний левый угол
			{
				test1 := cellInfo[y*int16(platform.Width)+x] == CELL_TYPE_SPACE
				test2 := (x == 0) || (cellInfo[y*int16(platform.Width)+(x-PLATFORM_BLOCK_SIZE_3x3)] == CELL_TYPE_BLOCK)
				test3 := (y == 0) || (cellInfo[(y-PLATFORM_BLOCK_SIZE_3x3)*int16(platform.Width)+x] == CELL_TYPE_BLOCK)

				if test1 && test2 && test3 {
					platform.Objects, _ = appendObjects(platform.Objects,
						platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_CORNER],
						float64(x), float64(y), 0, 3)

					cellsWalls[(y+0)*int16(platform.Width)+(x+0)] = CELL_TYPE_WALL
					cellsWalls[(y+1)*int16(platform.Width)+(x+0)] = CELL_TYPE_WALL
					cellsWalls[(y+2)*int16(platform.Width)+(x+0)] = CELL_TYPE_WALL
					cellsWalls[(y+0)*int16(platform.Width)+(x+1)] = CELL_TYPE_WALL
					cellsWalls[(y+0)*int16(platform.Width)+(x+2)] = CELL_TYPE_WALL
					continue
				}
			}
			// верхний правый угол
			{
				test1 := cellInfo[y*int16(platform.Width)+x] == CELL_TYPE_SPACE
				test2 := (x == PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_3x3) ||
					(cellInfo[y*int16(platform.Width)+(x+PLATFORM_BLOCK_SIZE_3x3)] == CELL_TYPE_BLOCK)
				test3 := (y == 0) ||
					(cellInfo[(y-PLATFORM_BLOCK_SIZE_3x3)*int16(platform.Width)+x] == CELL_TYPE_BLOCK)

				if test1 && test2 && test3 {
					platform.Objects, _ = appendObjects(platform.Objects,
						platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_CORNER],
						float64(x), float64(y),
						3, 3)

					cellsWalls[(y+0)*int16(platform.Width)+(x+2)] = CELL_TYPE_WALL
					cellsWalls[(y+1)*int16(platform.Width)+(x+2)] = CELL_TYPE_WALL
					cellsWalls[(y+2)*int16(platform.Width)+(x+2)] = CELL_TYPE_WALL
					cellsWalls[(y+0)*int16(platform.Width)+(x+0)] = CELL_TYPE_WALL
					cellsWalls[(y+0)*int16(platform.Width)+(x+1)] = CELL_TYPE_WALL

					continue
				}
			}
			// нижний правый угол
			{
				test1 := cellInfo[y*int16(platform.Width)+x] == CELL_TYPE_SPACE
				test2 := (x == PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_3x3) ||
					(cellInfo[y*int16(platform.Width)+(x+PLATFORM_BLOCK_SIZE_3x3)] == CELL_TYPE_BLOCK)
				test3 := (y == PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_3x3) ||
					(cellInfo[(y+PLATFORM_BLOCK_SIZE_3x3)*int16(platform.Width)+x] == CELL_TYPE_BLOCK)

				if test1 && test2 && test3 {
					platform.Objects, _ = appendObjects(platform.Objects,
						platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_CORNER],
						float64(x), float64(y),
						2, 3)

					cellsWalls[(y+0)*int16(platform.Width)+(x+2)] = CELL_TYPE_WALL
					cellsWalls[(y+1)*int16(platform.Width)+(x+2)] = CELL_TYPE_WALL
					cellsWalls[(y+2)*int16(platform.Width)+(x+2)] = CELL_TYPE_WALL
					cellsWalls[(y+2)*int16(platform.Width)+(x+0)] = CELL_TYPE_WALL
					cellsWalls[(y+2)*int16(platform.Width)+(x+1)] = CELL_TYPE_WALL

					continue
				}
			}
			// нижний левый угол
			{
				test1 := cellInfo[y*int16(platform.Width)+x] == CELL_TYPE_SPACE
				test2 := (x == 0) ||
					(cellInfo[y*int16(platform.Width)+(x-PLATFORM_BLOCK_SIZE_3x3)] == CELL_TYPE_BLOCK)
				test3 := (y == PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_3x3) ||
					(cellInfo[(y+PLATFORM_BLOCK_SIZE_3x3)*int16(platform.Width)+x] == CELL_TYPE_BLOCK)

				if test1 && test2 && test3 {
					platform.Objects, _ = appendObjects(platform.Objects,
						platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_CORNER],
						float64(x), float64(y),
						1, 3)

					cellsWalls[(y+0)*int16(platform.Width)+(x+0)] = CELL_TYPE_WALL
					cellsWalls[(y+1)*int16(platform.Width)+(x+0)] = CELL_TYPE_WALL
					cellsWalls[(y+2)*int16(platform.Width)+(x+0)] = CELL_TYPE_WALL
					cellsWalls[(y+2)*int16(platform.Width)+(x+1)] = CELL_TYPE_WALL
					cellsWalls[(y+2)*int16(platform.Width)+(x+2)] = CELL_TYPE_WALL

					continue
				}
			}
			// левая стена
			{
				test1 := cellInfo[y*int16(platform.Width)+x] == CELL_TYPE_SPACE
				test2 := (x == 0) ||
					(cellInfo[y*int16(platform.Width)+(x-PLATFORM_BLOCK_SIZE_3x3)] == CELL_TYPE_BLOCK)
				test3 := (y != 0) &&
					(cellInfo[(y-PLATFORM_BLOCK_SIZE_3x3)*int16(platform.Width)+x] == CELL_TYPE_SPACE)
				test4 := (y != PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_3x3) &&
					(cellInfo[(y+PLATFORM_BLOCK_SIZE_3x3)*int16(platform.Width)+x] == CELL_TYPE_SPACE)

				if test1 && test2 && test3 && test4 {
					platform.Objects, _ = appendObjects(platform.Objects,
						platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_WALL],
						float64(x), float64(y),
						0, 3)

					cellsWalls[(y+0)*int16(platform.Width)+x] = CELL_TYPE_WALL
					cellsWalls[(y+1)*int16(platform.Width)+x] = CELL_TYPE_WALL
					cellsWalls[(y+2)*int16(platform.Width)+x] = CELL_TYPE_WALL

					continue
				}
			}
			// правая стена
			{
				test1 := cellInfo[y*int16(platform.Width)+x] == CELL_TYPE_SPACE
				test2 := (x == PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_3x3) ||
					(cellInfo[y*int16(platform.Width)+(x+PLATFORM_BLOCK_SIZE_3x3)] == CELL_TYPE_BLOCK)
				test3 := (y != 0) &&
					(cellInfo[(y-PLATFORM_BLOCK_SIZE_3x3)*int16(platform.Width)+x] == CELL_TYPE_SPACE)
				test4 := (y != PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_3x3) &&
					(cellInfo[(y+PLATFORM_BLOCK_SIZE_3x3)*int16(platform.Width)+x] == CELL_TYPE_SPACE)

				if test1 && test2 && test3 && test4 {
					platform.Objects, _ = appendObjects(platform.Objects,
						platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_WALL],
						float64(x), float64(y),
						2, 3)

					cellsWalls[(y+0)*int16(platform.Width)+(x+2)] = CELL_TYPE_WALL
					cellsWalls[(y+1)*int16(platform.Width)+(x+2)] = CELL_TYPE_WALL
					cellsWalls[(y+2)*int16(platform.Width)+(x+2)] = CELL_TYPE_WALL

					continue
				}
			}
			// верхняя стена
			{
				test1 := cellInfo[y*int16(platform.Width)+x] == CELL_TYPE_SPACE
				test2 := (y == 0) ||
					(cellInfo[(y-PLATFORM_BLOCK_SIZE_3x3)*int16(platform.Width)+x] == CELL_TYPE_BLOCK)
				test3 := (x != 0) &&
					(cellInfo[y*int16(platform.Width)+(x-PLATFORM_BLOCK_SIZE_3x3)] == CELL_TYPE_SPACE)
				test4 := (x != PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_3x3) &&
					(cellInfo[y*int16(platform.Width)+(x+PLATFORM_BLOCK_SIZE_3x3)] == CELL_TYPE_SPACE)

				if test1 && test2 && test3 && test4 {
					platform.Objects, _ = appendObjects(platform.Objects,
						platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_WALL],
						float64(x), float64(y),
						3, 3)

					cellsWalls[(y+0)*int16(platform.Width)+(x+0)] = CELL_TYPE_WALL
					cellsWalls[(y+0)*int16(platform.Width)+(x+1)] = CELL_TYPE_WALL
					cellsWalls[(y+0)*int16(platform.Width)+(x+2)] = CELL_TYPE_WALL

					continue
				}
			}
			// нижняя стена
			{
				test1 := cellInfo[y*int16(platform.Width)+x] == CELL_TYPE_SPACE
				test2 := (y == PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_3x3) ||
					(cellInfo[(y+PLATFORM_BLOCK_SIZE_3x3)*int16(platform.Width)+x] == CELL_TYPE_BLOCK)
				test3 := (x != 0) &&
					(cellInfo[y*int16(platform.Width)+(x-PLATFORM_BLOCK_SIZE_3x3)] == CELL_TYPE_SPACE)
				test4 := (x != PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_3x3) &&
					(cellInfo[y*int16(platform.Width)+(x+PLATFORM_BLOCK_SIZE_3x3)] == CELL_TYPE_SPACE)

				if test1 && test2 && test3 && test4 {
					platform.Objects, _ = appendObjects(platform.Objects,
						platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_WALL],
						float64(x), float64(y),
						1, 3)

					cellsWalls[(y+2)*int16(platform.Width)+(x+0)] = CELL_TYPE_WALL
					cellsWalls[(y+2)*int16(platform.Width)+(x+1)] = CELL_TYPE_WALL
					cellsWalls[(y+2)*int16(platform.Width)+(x+2)] = CELL_TYPE_WALL
				}
			}

		}
	}
}

func createPlatformElements(platform *Platform, cellInfo []PlatformCellType) {
	empty := make([]Point16, 0)
	for y := int16(0); y < PLATFORM_WORK_SIZE; y += PLATFORM_BLOCK_SIZE_3x3 {
		for x := int16(0); x < PLATFORM_WORK_SIZE; x += PLATFORM_BLOCK_SIZE_3x3 {
			// если в центре дырка, то дальше от центра
			{
				test1 := platform.HaveDecor || cellInfo[PLATFORM_WORK_SIZE/2*platform.Width+PLATFORM_WORK_SIZE/2] == CELL_TYPE_PIT
				test2 := x > PLATFORM_BLOCK_SIZE_6x6 &&
					x < PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_6x6 &&
					y > PLATFORM_BLOCK_SIZE_6x6 &&
					y < PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_6x6
				test3 := y/PLATFORM_BLOCK_SIZE_6x6 == PLATFORM_WORK_SIZE/2/PLATFORM_BLOCK_SIZE_6x6 &&
					x/PLATFORM_BLOCK_SIZE_6x6 == PLATFORM_WORK_SIZE/2/PLATFORM_BLOCK_SIZE_6x6

				if test1 && (test2 || test3) {
					log.Printf("Continue1 at %d %d\n", x, y)
					continue
				}
			}

			{
				testPoint := NewPoint16(x, y)
				testPoint = testPoint.Div(PLATFORM_BLOCK_SIZE_6x6)

				test1 := getPortalCoord(DIR_NORTH, platform.ExitCoord)
				test1 = test1.Div(PLATFORM_BLOCK_SIZE_6x6)

				east := getPortalCoord(DIR_EAST, platform.ExitCoord)
				test2 := NewPoint16(east.X-PLATFORM_BLOCK_SIZE_6x6, east.Y)
				test2 = test2.Div(PLATFORM_BLOCK_SIZE_6x6)

				south := getPortalCoord(DIR_SOUTH, platform.ExitCoord)
				test3 := NewPoint16(south.X, south.Y-PLATFORM_BLOCK_SIZE_6x6)
				test3 = test3.Div(PLATFORM_BLOCK_SIZE_6x6)

				test4 := getPortalCoord(DIR_WEST, platform.ExitCoord)
				test4 = test4.Div(PLATFORM_BLOCK_SIZE_6x6)

				if (testPoint == test1) || (testPoint == test2) || (testPoint == test3) || (testPoint == test4) {
					log.Printf("Continue2 at %d %d\n", x, y)
					continue
				}
			}

			// есть проход
			if (cellInfo[(y+1)*int16(platform.Width)+(x+1)] & CELL_TYPE_WALK) == 0 {
				log.Printf("Continue3 at %d %d\n", x, y)
				continue
			}

			empty = append(empty, NewPoint16(x, y))
		}
	}

	// Shuffle
	for i := range empty {
		j := rand.Intn(i + 1)
		empty[i], empty[j] = empty[j], empty[i]
	}

	pills := rand.Int() % 5
	coffs := 1 + rand.Int()%2
	env := rand.Int()%3 + 1

	is := 0
	if len(empty) < (pills + env) {
		is = len(empty)
	} else {
		is = pills + env
	}

	for i := 0; i < is; i++ {
		if coffs > 0 {
			if createCoffins(platform, empty[i], cellInfo) {
				coffs--
				continue
			}
		}
		if pills > 0 { // столбы
			if createPillars(platform, empty[i], cellInfo) {
				pills--
				continue
			}
		}
		if env > 0 { // свечи
			if createEnvironment(platform, empty[i], cellInfo) {
				env--
				continue
			}
		}
	}
}

func createCoffins(platform *Platform, point Point16, cellInfo []PlatformCellType) bool {
	w := int16(platform.Width)
	x := point.X
	y := point.Y

	// TODO: Test
	if (y < PLATFORM_BLOCK_SIZE_3x3) || (x < PLATFORM_BLOCK_SIZE_3x3) {
		return false
	}

	// есть проходы рядом
	if (cellInfo[y*w+x-PLATFORM_BLOCK_SIZE_3x3]&CELL_TYPE_WALK) == 0 ||
		(cellInfo[y*w+x+PLATFORM_BLOCK_SIZE_3x3]&CELL_TYPE_WALK) == 0 ||
		(cellInfo[(y+PLATFORM_BLOCK_SIZE_3x3)*w+x]&CELL_TYPE_WALK) == 0 ||
		(cellInfo[(y-PLATFORM_BLOCK_SIZE_3x3)*w+x]&CELL_TYPE_WALK) == 0 ||
		(cellInfo[(y-PLATFORM_BLOCK_SIZE_3x3)*w+x-PLATFORM_BLOCK_SIZE_3x3]&CELL_TYPE_WALK) == 0 ||
		(cellInfo[(y-PLATFORM_BLOCK_SIZE_3x3)*w+x+PLATFORM_BLOCK_SIZE_3x3]&CELL_TYPE_WALK) == 0 ||
		(cellInfo[(y+PLATFORM_BLOCK_SIZE_3x3)*w+x-PLATFORM_BLOCK_SIZE_3x3]&CELL_TYPE_WALK) == 0 ||
		(cellInfo[(y+PLATFORM_BLOCK_SIZE_3x3)*w+x+PLATFORM_BLOCK_SIZE_3x3]&CELL_TYPE_WALK) == 0 {
		// ничего не делаем
		return false
	} else {
		if len(platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_COFFIN]) > 0 {
			objects, item := appendObjects(platform.Objects,
				platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_COFFIN],
				float64(x), float64(y),
				int8(rand.Int()%4),
				1.0)
			platform.Objects = objects

			for yy := int16(0); yy < item.Width; yy++ {
				for xx := int16(0); xx < item.Height; xx++ {
					cellInfo[(y+yy)*w+(x+xx)] = CELL_TYPE_WALL
				}
			}
		}
		return true
	}
	return false
}

func createPillars(platform *Platform, point Point16, cellInfo []PlatformCellType) bool {
	x := point.X
	y := point.Y
	w := int16(platform.Width)
	ww := int16(PLATFORM_WORK_SIZE)
	hh := int16(PLATFORM_WORK_SIZE)

	// TODO: Test
	if (y < PLATFORM_BLOCK_SIZE_3x3) || (x < PLATFORM_BLOCK_SIZE_3x3) {
		return false
	}

	// есть проходы рядом
	if x < PLATFORM_BLOCK_SIZE_6x6 || x > ww-PLATFORM_BLOCK_SIZE_6x6 || y < PLATFORM_BLOCK_SIZE_6x6 || y > hh-PLATFORM_BLOCK_SIZE_6x6 ||
		(cellInfo[y*w+x-PLATFORM_BLOCK_SIZE_3x3]&CELL_TYPE_WALK) == 0 ||
		(cellInfo[y*w+x+PLATFORM_BLOCK_SIZE_3x3]&CELL_TYPE_WALK) == 0 ||
		(cellInfo[(y+PLATFORM_BLOCK_SIZE_3x3)*w+x]&CELL_TYPE_WALK) == 0 ||
		(cellInfo[(y-PLATFORM_BLOCK_SIZE_3x3)*w+x]&CELL_TYPE_WALK) == 0 {
		// ничего не делаем
		return false
	} else {
		platform.Objects, _ = appendObjects(platform.Objects,
			platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_PILLAR],
			float64(x), float64(y), int8(rand.Int()%4), 2.0)
		cellInfo[(y+0)*w+x+0] = CELL_TYPE_WALL
		cellInfo[(y+0)*w+x+1] = CELL_TYPE_WALL
		cellInfo[(y+1)*w+x+0] = CELL_TYPE_WALL
		cellInfo[(y+1)*w+x+1] = CELL_TYPE_WALL
		return true
	}
	return false
}

func createEnvironment(platform *Platform, point Point16, cellInfo []PlatformCellType) bool {
	x := point.X
	y := point.Y
	w := int16(platform.Width)
	ww := int16(PLATFORM_WORK_SIZE)
	hh := int16(PLATFORM_WORK_SIZE)

	// TODO: Test
	if (y < PLATFORM_BLOCK_SIZE_3x3) || (x < PLATFORM_BLOCK_SIZE_3x3) {
		return false
	}

	// дальше от центра
	if (x > PLATFORM_BLOCK_SIZE_6x6 && x < ww-PLATFORM_BLOCK_SIZE_6x6 && y > PLATFORM_BLOCK_SIZE_6x6 && y < hh-PLATFORM_BLOCK_SIZE_6x6) ||
		(y/PLATFORM_BLOCK_SIZE_6x6 == hh/2/PLATFORM_BLOCK_SIZE_6x6 && x/PLATFORM_BLOCK_SIZE_6x6 == ww/2/PLATFORM_BLOCK_SIZE_6x6) {
		// ничего не делаем
		return false
	} else {
		nearWall := false
		offset := NewPointFloat(0.0, 0.0)
		if (cellInfo[(y+1)*w+x] & CELL_TYPE_WALK) == 0 {
			offset.X += 1.0
			nearWall = true
		} else if (cellInfo[(y+1)*w+(x+2)] & CELL_TYPE_WALK) == 0 {
			nearWall = true
		}

		if (cellInfo[y*w+(x+1)] & CELL_TYPE_WALK) == 0 {
			offset.Y += 1.0
			nearWall = true
		} else if (cellInfo[(y+2)*w+(x+1)] & CELL_TYPE_WALK) == 0 {
			nearWall = true
		}

		if nearWall == false {
			offset.X += 0.5
			offset.Y += 0.5
			cellInfo[(y+1)*w+x+1] = CELL_TYPE_WALL
		}

		x = x + int16(offset.X)
		y = y + int16(offset.Y)

		platform.Objects, _ = appendObjects(platform.Objects,
			platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_ENVIRONMENT],
			float64(x), float64(y), 0,
			3)
		return true
	}
	return false
}

// TODO: В качестве параметра float x,y???
func appendObjects(container []PlatformObject, objects []*PlatformObjectInfo, x, y float64, rot int8, size int16) ([]PlatformObject, *PlatformObjectInfo) {
	if len(objects) == 0 {
		return container, nil
	}

	var selectedItem *PlatformObjectInfo = nil

	// Random probability
	sumProb := 0
	for i := range objects {
		sumProb += int(objects[i].Probability * 100)
	}
	randVal := rand.Int() % sumProb

	// Select random item
	variant := 0
	for i := range objects {
		selected := (variant <= randVal) && (variant+int(objects[i].Probability*100) > randVal)
		if selected {
			selectedItem = objects[i]
			break
		} else {
			variant += int(objects[i].Probability * 100)
		}
	}

	if selectedItem == nil {
		return container, nil
	}

	// Max size
	maxInt16 := func(a, b int16) int16 {
		if a > b {
			return a
		}
		return b
	}
	size = maxInt16(maxInt16(selectedItem.Width, selectedItem.Height), size)

	// Position
	if rot == 1 {
		y += float64(size)
	} else if rot == 2 {
		x += float64(size)
		y += float64(size)
	} else if rot == 3 {
		x += float64(size)
	}

	// Append
	object := PlatformObject{
		Id:  selectedItem.Id,
		X:   x,
		Y:   y,
		Rot: rot,
	}

	container = append(container, object)
	return container, selectedItem
}
