package polytopiamapmodel

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	fileOffsetMap = make(map[string]int)
	DebugMode     = false // Set to true to enable debug output
)

// debugPrint prints a message only if debug mode is enabled
func debugPrint(format string, args ...interface{}) {
	if DebugMode {
		fmt.Printf(format, args...)
	}
}

type PolytopiaSaveOutput struct {
	MapHeight         int
	MapWidth          int
	GameVersion       int
	MapHeaderOutput   MapHeaderOutput
	InitialTileData   [][]TileData
	InitialPlayerData []PlayerData
	TileData          [][]TileData
	MaxTurn           int
	PlayerData        []PlayerData
	FileOffsetMap     map[string]int
	OwnerTribeMap     map[int]int
	TribeCityMap      map[int][]CityLocationData
	TurnCaptureMap    map[int][]ActionCaptureCity
}

type CityLocationData struct {
	X        int
	Y        int
	CityName string
	Capital  int
}

func convertByteListToInt(oldArr []byte) []int {
	newArr := make([]int, len(oldArr))
	for i := 0; i < len(newArr); i++ {
		newArr[i] = int(oldArr[i])
	}
	return newArr
}

func readTileData(streamReader *io.SectionReader, tileData [][]TileData, mapWidth int, mapHeight int, gameVersion int) {
	updateFileOffsetMap(fileOffsetMap, streamReader, buildMapStartKey())

	for i := 0; i < int(mapHeight); i++ {
		for j := 0; j < int(mapWidth); j++ {
			tileStartKey := buildTileStartKey(j, i)
			updateFileOffsetMap(fileOffsetMap, streamReader, tileStartKey)

			tileData[i][j] = DeserializeTileDataFromBytes(streamReader, i, j, gameVersion)

			tileEndKey := buildTileEndKey(j, i)
			updateFileOffsetMap(fileOffsetMap, streamReader, tileEndKey)
		}
	}

	updateFileOffsetMap(fileOffsetMap, streamReader, buildMapEndKey())
}

func readAllPlayerData(streamReader *io.SectionReader, gameVersion int) []PlayerData {
	allPlayersStartKey := buildAllPlayersStartKey()
	updateFileOffsetMap(fileOffsetMap, streamReader, allPlayersStartKey)

	numPlayers, _ := readUint16Safe(streamReader, "number of players")

	allPlayerData := make([]PlayerData, int(numPlayers))

	for i := 0; i < int(numPlayers); i++ {
		debugPrint("  Reading player %d/%d...\n", i+1, numPlayers)
		playerData := DeserializePlayerDataFromBytes(streamReader, gameVersion)
		allPlayerData[i] = playerData
		debugPrint("  Player %d read - Name: %s, Tribe: %d\n", i+1, playerData.Name, playerData.Tribe)
	}

	allPlayersEndKey := buildAllPlayersEndKey()
	updateFileOffsetMap(fileOffsetMap, streamReader, allPlayersEndKey)

	return allPlayerData
}

func buildOwnerTribeMap(allPlayerData []PlayerData) map[int]int {
	ownerTribeMap := make(map[int]int)

	for i := 0; i < len(allPlayerData); i++ {
		playerData := allPlayerData[i]
		mappedTribe, ok := ownerTribeMap[playerData.PlayerId]
		if ok {
			log.Fatal(fmt.Sprintf("Owner to tribe map has duplicate player id %v already mapped to %v", playerData.PlayerId, mappedTribe))
		}
		ownerTribeMap[playerData.PlayerId] = playerData.Tribe
	}

	return ownerTribeMap
}

func buildTribeCityMap(currentMapHeaderOutput MapHeaderOutput, tileData [][]TileData) map[int][]CityLocationData {
	tribeCityMap := make(map[int][]CityLocationData)
	for i := 0; i < int(currentMapHeaderOutput.MapHeight); i++ {
		for j := 0; j < int(currentMapHeaderOutput.MapWidth); j++ {
			if tileData[i][j].ImprovementData != nil && tileData[i][j].ImprovementType == 1 {
				tribeOwner := tileData[i][j].Owner
				_, ok := tribeCityMap[tribeOwner]
				if !ok {
					tribeCityMap[tribeOwner] = make([]CityLocationData, 0)
				}

				cityName := ""
				if tileData[i][j].ImprovementData != nil {
					cityName = tileData[i][j].ImprovementData.CityName
				}

				cityLocationData := CityLocationData{
					X:        tileData[i][j].WorldCoordinates[0],
					Y:        tileData[i][j].WorldCoordinates[1],
					CityName: cityName,
					Capital:  tileData[i][j].Capital,
				}
				tribeCityMap[tribeOwner] = append(tribeCityMap[tribeOwner], cityLocationData)
			}
		}
	}
	return tribeCityMap
}

// Read compressed .state file without generating a decompressed file
// Can be used with read only applications to display map data
func ReadPolytopiaCompressedFile(inputFilename string) (*PolytopiaSaveOutput, error) {
	decompressedReader, decompressedLength := BuildReaderForDecompressedFile(inputFilename)
	streamReader := io.NewSectionReader(decompressedReader, int64(0), int64(decompressedLength))

	return ParsePolytopiaFile(streamReader)
}

// Read decompressed file
// Should be used with applications that need to modify decompressed data directly
func ReadPolytopiaDecompressedFile(inputFilename string) (*PolytopiaSaveOutput, error) {
	inputFile, err := os.OpenFile(inputFilename, os.O_RDWR, 0644)
	defer inputFile.Close()
	if err != nil {
		log.Fatal("Failed to load save state: ", err)
		return nil, err
	}
	fi, err := inputFile.Stat()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	fileLength := fi.Size()
	streamReader := io.NewSectionReader(inputFile, int64(0), fileLength)

	return ParsePolytopiaFile(streamReader)
}

func ParsePolytopiaFile(streamReader *io.SectionReader) (*PolytopiaSaveOutput, error) {
	fileOffsetMap = make(map[string]int)

	// Read initial map state
	debugPrint("Reading initial map header...\n")
	updateFileOffsetMap(fileOffsetMap, streamReader, buildMapHeaderStartKey())
	initialMapHeaderOutput := DeserializeMapHeaderFromBytes(streamReader)
	updateFileOffsetMap(fileOffsetMap, streamReader, buildMapHeaderEndKey())
	debugPrint("Initial map header read - Size: %dx%d, Version: %d\n",
		initialMapHeaderOutput.MapWidth, initialMapHeaderOutput.MapHeight,
		initialMapHeaderOutput.MapHeaderInput.Version1)

	initialTileData := make([][]TileData, initialMapHeaderOutput.MapHeight)
	for i := 0; i < initialMapHeaderOutput.MapHeight; i++ {
		initialTileData[i] = make([]TileData, initialMapHeaderOutput.MapWidth)
	}
	gameVersion := int(initialMapHeaderOutput.MapHeaderInput.Version1)
	debugPrint("Reading initial tile data...\n")
	readTileData(streamReader, initialTileData, initialMapHeaderOutput.MapWidth, initialMapHeaderOutput.MapHeight, gameVersion)
	debugPrint("Reading initial player data...\n")
	initialPlayerData := readAllPlayerData(streamReader, gameVersion)
	debugPrint("Initial player data read - %d players\n", len(initialPlayerData))

	ownerTribeMap := buildOwnerTribeMap(initialPlayerData)

	_ = readFixedList(streamReader, 3)

	// Read current map state
	debugPrint("Reading current map header...\n")
	updateFileOffsetMap(fileOffsetMap, streamReader, buildMapHeaderStartKey())
	currentMapHeaderOutput := DeserializeMapHeaderFromBytes(streamReader)
	updateFileOffsetMap(fileOffsetMap, streamReader, buildMapHeaderEndKey())
	debugPrint("Current map header read - Size: %dx%d\n",
		currentMapHeaderOutput.MapWidth, currentMapHeaderOutput.MapHeight)

	tileData := make([][]TileData, currentMapHeaderOutput.MapHeight)
	for i := 0; i < currentMapHeaderOutput.MapHeight; i++ {
		tileData[i] = make([]TileData, currentMapHeaderOutput.MapWidth)
	}
	debugPrint("Reading current tile data...\n")
	readTileData(streamReader, tileData, currentMapHeaderOutput.MapWidth, currentMapHeaderOutput.MapHeight, gameVersion)
	debugPrint("Reading current player data...\n")
	playerData := readAllPlayerData(streamReader, gameVersion)
	debugPrint("Current player data read - %d players\n", len(playerData))

	ownerTribeMap = buildOwnerTribeMap(playerData)
	tribeCityMap := buildTribeCityMap(currentMapHeaderOutput, tileData)

	_ = readFixedList(streamReader, 2)

	debugPrint("Reading actions...\n")
	turnCaptureMap := readAllActions(streamReader)
	debugPrint("Actions read - %d turns with captures\n", len(turnCaptureMap))

	output := &PolytopiaSaveOutput{
		MapHeight:         currentMapHeaderOutput.MapHeight,
		MapWidth:          currentMapHeaderOutput.MapWidth,
		GameVersion:       int(gameVersion),
		MapHeaderOutput:   currentMapHeaderOutput,
		InitialTileData:   initialTileData,
		InitialPlayerData: initialPlayerData,
		TileData:          tileData,
		MaxTurn:           int(currentMapHeaderOutput.MapHeaderInput.CurrentTurn),
		PlayerData:        playerData,
		FileOffsetMap:     fileOffsetMap,
		OwnerTribeMap:     ownerTribeMap,
		TribeCityMap:      tribeCityMap,
		TurnCaptureMap:    turnCaptureMap,
	}
	return output, nil
}
