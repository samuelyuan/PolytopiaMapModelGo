package polytopiamapmodel

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	fileOffsetMap = make(map[string]int)
)

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

func readAllPlayerData(streamReader *io.SectionReader) []PlayerData {
	allPlayersStartKey := buildAllPlayersStartKey()
	updateFileOffsetMap(fileOffsetMap, streamReader, allPlayersStartKey)

	numPlayers := unsafeReadUint16(streamReader)
	allPlayerData := make([]PlayerData, int(numPlayers))

	for i := 0; i < int(numPlayers); i++ {
		playerData := DeserializePlayerDataFromBytes(streamReader)
		allPlayerData[i] = playerData
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
	updateFileOffsetMap(fileOffsetMap, streamReader, buildMapHeaderStartKey())
	initialMapHeaderOutput := DeserializeMapHeaderFromBytes(streamReader)
	updateFileOffsetMap(fileOffsetMap, streamReader, buildMapHeaderEndKey())

	initialTileData := make([][]TileData, initialMapHeaderOutput.MapHeight)
	for i := 0; i < initialMapHeaderOutput.MapHeight; i++ {
		initialTileData[i] = make([]TileData, initialMapHeaderOutput.MapWidth)
	}
	gameVersion := int(initialMapHeaderOutput.MapHeaderInput.Version1)
	readTileData(streamReader, initialTileData, initialMapHeaderOutput.MapWidth, initialMapHeaderOutput.MapHeight, gameVersion)
	initialPlayerData := readAllPlayerData(streamReader)

	ownerTribeMap := buildOwnerTribeMap(initialPlayerData)

	_ = readFixedList(streamReader, 3)

	// Read current map state
	updateFileOffsetMap(fileOffsetMap, streamReader, buildMapHeaderStartKey())
	currentMapHeaderOutput := DeserializeMapHeaderFromBytes(streamReader)
	updateFileOffsetMap(fileOffsetMap, streamReader, buildMapHeaderEndKey())

	tileData := make([][]TileData, currentMapHeaderOutput.MapHeight)
	for i := 0; i < currentMapHeaderOutput.MapHeight; i++ {
		tileData[i] = make([]TileData, currentMapHeaderOutput.MapWidth)
	}
	readTileData(streamReader, tileData, currentMapHeaderOutput.MapWidth, currentMapHeaderOutput.MapHeight, gameVersion)
	playerData := readAllPlayerData(streamReader)

	ownerTribeMap = buildOwnerTribeMap(playerData)
	tribeCityMap := buildTribeCityMap(currentMapHeaderOutput, tileData)

	_ = readFixedList(streamReader, 2)

	turnCaptureMap := readAllActions(streamReader)

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
