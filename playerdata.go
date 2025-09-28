package polytopiamapmodel

import (
	"encoding/binary"
	"fmt"
	"image/color"
	"io"
	"log"
)

type CityLocationData struct {
	X        int
	Y        int
	CityName string
	Capital  int
}

type PlayerData struct {
	PlayerId             int
	Name                 string
	AccountId            string
	AutoPlay             bool
	StartTileCoordinates [2]int
	Tribe                int
	UnknownByte1         int
	DifficultyHandicap   int
	AggressionsByPlayers []PlayerAggression
	Currency             int
	Score                int
	UnknownInt2          int
	NumCities            int
	AvailableTech        []int
	EncounteredPlayers   []int
	Tasks                []PlayerTaskData
	TotalUnitsKilled     int
	TotalUnitsLost       int
	TotalTribesDestroyed int
	OverrideColor        []int
	OverrideTribe        byte
	UniqueImprovements   []int
	DiplomacyArr         []DiplomacyData
	DiplomacyMessages    []DiplomacyMessage
	DestroyedByTribe     int
	DestroyedTurn        int
	UnknownBuffer2       []int
	EndScore             int
	PlayerSkin           int
	UnknownBuffer3       []int
}

type PlayerAggression struct {
	PlayerId   int
	Aggression int
}

type PlayerTaskData struct {
	Type   int
	Buffer []int
}

type DiplomacyMessage struct {
	MessageType int
	Sender      int
}

type DiplomacyData struct {
	PlayerId               uint8
	DiplomacyRelationState uint8
	LastAttackTurn         int32
	EmbassyLevel           uint8
	LastPeaceBrokenTurn    int32
	FirstMeet              int32
	EmbassyBuildTurn       int32
	PreviousAttackTurn     int32
}

func DeserializePlayerDataFromBytes(streamReader *io.SectionReader, gameVersion int) PlayerData {
	playerId, _ := readUint8Safe(streamReader, "player ID")
	playerName := readVarString(streamReader, "playerName")
	playerAccountId := readVarString(streamReader, "playerAccountId")
	autoPlay, _ := readUint8Safe(streamReader, "auto play flag")
	startTileCoordinates1, _ := readInt32Safe(streamReader, "start coordinates 1")
	startTileCoordinates2, _ := readInt32Safe(streamReader, "start coordinates 2")
	tribe, _ := readUint16Safe(streamReader, "tribe")
	unknownByte1, _ := readUint8Safe(streamReader, "unknown byte")
	difficultyHandicap, _ := readUint32Safe(streamReader, "difficulty handicap")

	var aggressionsByPlayers []PlayerAggression
	if gameVersion < 114 {
		unknownArrLen1, _ := readUint16Safe(streamReader, "aggressions array length")
		aggressionsByPlayers = make([]PlayerAggression, 0)
		for i := 0; i < int(unknownArrLen1); i++ {
			playerIdOther, _ := readUint8Safe(streamReader, fmt.Sprintf("aggression player ID %d", i))
			aggression, _ := readInt32Safe(streamReader, fmt.Sprintf("aggression value %d", i))
			aggressionsByPlayers = append(aggressionsByPlayers, PlayerAggression{
				PlayerId:   int(playerIdOther),
				Aggression: int(aggression),
			})
		}
	} else {
		debugPrint("    Skipping aggressions array for version %d\n", gameVersion)
		aggressionsByPlayers = make([]PlayerAggression, 0)
	}

	currency, _ := readUint32Safe(streamReader, "currency")
	score, _ := readUint32Safe(streamReader, "score")
	unknownInt2, _ := readUint32Safe(streamReader, "unknown int")
	numCities, _ := readUint16Safe(streamReader, "number of cities")

	techArrayLen, _ := readUint16Safe(streamReader, "tech array length")
	techArray := make([]int, techArrayLen)
	for i := 0; i < int(techArrayLen); i++ {
		techType := unsafeReadUint16(streamReader)
		techArray[i] = int(techType)
	}

	encounteredPlayersLen, _ := readUint16Safe(streamReader, "encountered players length")
	encounteredPlayers := make([]int, 0)
	for i := 0; i < int(encounteredPlayersLen); i++ {
		playerId := unsafeReadUint8(streamReader)
		encounteredPlayers = append(encounteredPlayers, int(playerId))
	}

	numTasks, _ := readInt16Safe(streamReader, "number of tasks")
	taskArr := make([]PlayerTaskData, int(numTasks))
	for i := 0; i < int(numTasks); i++ {
		taskType := unsafeReadInt16(streamReader)

		var buffer []byte
		if taskType == 1 || taskType == 5 { // Task type 1 is Pacifist, type 5 is Killer
			buffer = readFixedList(streamReader, 6) // Extra buffer contains a uint32
		} else if taskType >= 1 && taskType <= 8 {
			buffer = readFixedList(streamReader, 2)
		} else {
			log.Fatal("Invalid task type:", taskType)
		}
		taskArr[i] = PlayerTaskData{
			Type:   int(taskType),
			Buffer: convertByteListToInt(buffer),
		}
	}

	totalKills, _ := readInt32Safe(streamReader, "total kills")
	totalLosses, _ := readInt32Safe(streamReader, "total losses")
	totalTribesDestroyed, _ := readInt32Safe(streamReader, "total tribes destroyed")
	debugPrint("    Reading override color...\n")
	overrideColor := convertByteListToInt(readFixedList(streamReader, 4))
	overrideTribe, _ := readUint8Safe(streamReader, "override tribe")

	playerUniqueImprovementsSize, _ := readUint16Safe(streamReader, "player unique improvements size")
	playerUniqueImprovements := make([]int, int(playerUniqueImprovementsSize))
	for i := 0; i < int(playerUniqueImprovementsSize); i++ {
		improvement := unsafeReadUint16(streamReader)
		playerUniqueImprovements[i] = int(improvement)
	}

	diplomacyArrLen, _ := readUint16Safe(streamReader, "diplomacy array length")
	diplomacyArr := make([]DiplomacyData, int(diplomacyArrLen))
	for i := 0; i < len(diplomacyArr); i++ {
		diplomacyData := DiplomacyData{}
		if err := binary.Read(streamReader, binary.LittleEndian, &diplomacyData); err != nil {
			log.Fatal("Failed to load diplomacyData: ", err)
		}
		diplomacyArr[i] = diplomacyData
	}

	diplomacyMessagesSize, _ := readUint16Safe(streamReader, "diplomacy messages size")
	diplomacyMessagesArr := make([]DiplomacyMessage, int(diplomacyMessagesSize))
	for i := 0; i < int(diplomacyMessagesSize); i++ {
		messageType := unsafeReadUint8(streamReader)
		sender := unsafeReadUint8(streamReader)

		diplomacyMessagesArr[i] = DiplomacyMessage{
			MessageType: int(messageType),
			Sender:      int(sender),
		}
	}

	destroyedByTribe, _ := readUint8Safe(streamReader, "destroyed by tribe")
	destroyedTurn, _ := readUint32Safe(streamReader, "destroyed turn")
	debugPrint("    Reading unknown buffer 2...\n")
	unknownBuffer2 := convertByteListToInt(readFixedList(streamReader, 4))
	endScore, _ := readInt32Safe(streamReader, "end score")
	playerSkin, _ := readUint16Safe(streamReader, "player skin")
	debugPrint("    Reading unknown buffer 3...\n")
	unknownBuffer3 := convertByteListToInt(readFixedList(streamReader, 4))

	return PlayerData{
		PlayerId:             int(playerId),
		Name:                 playerName,
		AccountId:            playerAccountId,
		AutoPlay:             int(autoPlay) != 0,
		StartTileCoordinates: [2]int{int(startTileCoordinates1), int(startTileCoordinates2)},
		Tribe:                int(tribe),
		UnknownByte1:         int(unknownByte1),
		DifficultyHandicap:   int(difficultyHandicap),
		AggressionsByPlayers: aggressionsByPlayers,
		Currency:             int(currency),
		Score:                int(score),
		UnknownInt2:          int(unknownInt2),
		NumCities:            int(numCities),
		AvailableTech:        techArray,
		EncounteredPlayers:   encounteredPlayers,
		Tasks:                taskArr,
		TotalUnitsKilled:     int(totalKills),
		TotalUnitsLost:       int(totalLosses),
		TotalTribesDestroyed: int(totalTribesDestroyed),
		OverrideColor:        overrideColor,
		OverrideTribe:        overrideTribe,
		UniqueImprovements:   playerUniqueImprovements,
		DiplomacyArr:         diplomacyArr,
		DiplomacyMessages:    diplomacyMessagesArr,
		DestroyedByTribe:     int(destroyedByTribe),
		DestroyedTurn:        int(destroyedTurn),
		UnknownBuffer2:       unknownBuffer2,
		EndScore:             int(endScore),
		PlayerSkin:           int(playerSkin),
		UnknownBuffer3:       unknownBuffer3,
	}
}

func SerializePlayerDataToBytes(playerData PlayerData, gameVersion int) []byte {
	allPlayerData := make([]byte, 0)

	allPlayerData = append(allPlayerData, byte(playerData.PlayerId))
	allPlayerData = append(allPlayerData, ConvertVarString(playerData.Name)...)
	allPlayerData = append(allPlayerData, ConvertVarString(playerData.AccountId)...)
	allPlayerData = append(allPlayerData, ConvertBoolToByte(playerData.AutoPlay))
	allPlayerData = append(allPlayerData, ConvertUint32Bytes(playerData.StartTileCoordinates[0])...)
	allPlayerData = append(allPlayerData, ConvertUint32Bytes(playerData.StartTileCoordinates[1])...)
	allPlayerData = append(allPlayerData, ConvertUint16Bytes(playerData.Tribe)...)
	allPlayerData = append(allPlayerData, byte(playerData.UnknownByte1))
	allPlayerData = append(allPlayerData, ConvertUint32Bytes(playerData.DifficultyHandicap)...)

	// Only write aggressions array for versions < 114
	if gameVersion < 114 {
		allPlayerData = append(allPlayerData, ConvertUint16Bytes(len(playerData.AggressionsByPlayers))...)
		for i := 0; i < len(playerData.AggressionsByPlayers); i++ {
			allPlayerData = append(allPlayerData, byte(playerData.AggressionsByPlayers[i].PlayerId))
			allPlayerData = append(allPlayerData, ConvertUint32Bytes(playerData.AggressionsByPlayers[i].Aggression)...)
		}
	}

	allPlayerData = append(allPlayerData, ConvertUint32Bytes(playerData.Currency)...)
	allPlayerData = append(allPlayerData, ConvertUint32Bytes(playerData.Score)...)
	allPlayerData = append(allPlayerData, ConvertUint32Bytes(playerData.UnknownInt2)...)
	allPlayerData = append(allPlayerData, ConvertUint16Bytes(playerData.NumCities)...)

	allPlayerData = append(allPlayerData, ConvertUint16Bytes(len(playerData.AvailableTech))...)
	for i := 0; i < len(playerData.AvailableTech); i++ {
		allPlayerData = append(allPlayerData, ConvertUint16Bytes(playerData.AvailableTech[i])...)
	}

	allPlayerData = append(allPlayerData, ConvertUint16Bytes(len(playerData.EncounteredPlayers))...)
	for i := 0; i < len(playerData.EncounteredPlayers); i++ {
		allPlayerData = append(allPlayerData, byte(playerData.EncounteredPlayers[i]))
	}

	allPlayerData = append(allPlayerData, ConvertUint16Bytes(len(playerData.Tasks))...)
	for i := 0; i < len(playerData.Tasks); i++ {
		allPlayerData = append(allPlayerData, ConvertUint16Bytes(playerData.Tasks[i].Type)...)
		allPlayerData = append(allPlayerData, ConvertByteList(playerData.Tasks[i].Buffer)...)
	}

	allPlayerData = append(allPlayerData, ConvertUint32Bytes(playerData.TotalUnitsKilled)...)
	allPlayerData = append(allPlayerData, ConvertUint32Bytes(playerData.TotalUnitsLost)...)
	allPlayerData = append(allPlayerData, ConvertUint32Bytes(playerData.TotalTribesDestroyed)...)
	allPlayerData = append(allPlayerData, ConvertByteList(playerData.OverrideColor)...)

	allPlayerData = append(allPlayerData, playerData.OverrideTribe)

	allPlayerData = append(allPlayerData, ConvertUint16Bytes(len(playerData.UniqueImprovements))...)
	for i := 0; i < len(playerData.UniqueImprovements); i++ {
		allPlayerData = append(allPlayerData, ConvertUint16Bytes(playerData.UniqueImprovements[i])...)
	}

	allPlayerData = append(allPlayerData, ConvertUint16Bytes(len(playerData.DiplomacyArr))...)
	for i := 0; i < len(playerData.DiplomacyArr); i++ {
		allPlayerData = append(allPlayerData, SerializeDiplomacyDataToBytes(playerData.DiplomacyArr[i])...)
	}

	allPlayerData = append(allPlayerData, ConvertUint16Bytes(len(playerData.DiplomacyMessages))...)
	for i := 0; i < len(playerData.DiplomacyMessages); i++ {
		allPlayerData = append(allPlayerData, byte(playerData.DiplomacyMessages[i].MessageType))
		allPlayerData = append(allPlayerData, byte(playerData.DiplomacyMessages[i].Sender))
	}

	allPlayerData = append(allPlayerData, byte(playerData.DestroyedByTribe))
	allPlayerData = append(allPlayerData, ConvertUint32Bytes(playerData.DestroyedTurn)...)
	allPlayerData = append(allPlayerData, ConvertByteList(playerData.UnknownBuffer2)...)
	allPlayerData = append(allPlayerData, ConvertUint32Bytes(playerData.EndScore)...)
	allPlayerData = append(allPlayerData, ConvertUint16Bytes(playerData.PlayerSkin)...)
	allPlayerData = append(allPlayerData, ConvertByteList(playerData.UnknownBuffer3)...)

	return allPlayerData
}

func SerializeDiplomacyDataToBytes(diplomacyData DiplomacyData) []byte {
	data := make([]byte, 0)
	data = append(data, byte(diplomacyData.PlayerId))
	data = append(data, byte(diplomacyData.DiplomacyRelationState))
	data = append(data, ConvertUint32Bytes(int(diplomacyData.LastAttackTurn))...)
	data = append(data, byte(diplomacyData.EmbassyLevel))
	data = append(data, ConvertUint32Bytes(int(diplomacyData.LastPeaceBrokenTurn))...)
	data = append(data, ConvertUint32Bytes(int(diplomacyData.FirstMeet))...)
	data = append(data, ConvertUint32Bytes(int(diplomacyData.EmbassyBuildTurn))...)
	data = append(data, ConvertUint32Bytes(int(diplomacyData.PreviousAttackTurn))...)
	return data
}

func BuildEmptyPlayer(index int, playerName string, overrideColor color.RGBA) PlayerData {
	if index >= 254 {
		log.Fatal("Over 255 players")
	}

	// unknown array
	newArraySize := index + 1
	aggressionsByPlayers := make([]PlayerAggression, 0)
	for i := 1; i <= int(newArraySize); i++ {
		playerId := i
		if i == newArraySize {
			playerId = 255
		}
		aggressionsByPlayers = append(aggressionsByPlayers, PlayerAggression{
			PlayerId:   playerId,
			Aggression: 0,
		})
	}

	playerData := PlayerData{
		PlayerId:             index,
		Name:                 playerName,
		AccountId:            "00000000-0000-0000-0000-000000000000",
		AutoPlay:             true,
		StartTileCoordinates: [2]int{0, 0},
		Tribe:                2, // Ai-mo
		UnknownByte1:         1,
		DifficultyHandicap:   2,
		AggressionsByPlayers: aggressionsByPlayers,
		Currency:             5,
		Score:                0,
		UnknownInt2:          0,
		NumCities:            1,
		AvailableTech:        []int{},
		EncounteredPlayers:   []int{},
		Tasks:                []PlayerTaskData{},
		TotalUnitsKilled:     0,
		TotalUnitsLost:       0,
		TotalTribesDestroyed: 0,
		OverrideColor:        []int{int(overrideColor.B), int(overrideColor.G), int(overrideColor.R), 0},
		OverrideTribe:        0,
		UniqueImprovements:   []int{},
		DiplomacyArr:         []DiplomacyData{},
		DiplomacyMessages:    []DiplomacyMessage{},
		DestroyedByTribe:     0,
		DestroyedTurn:        0,
		UnknownBuffer2:       []int{255, 255, 255, 255},
		EndScore:             -1,
		PlayerSkin:           0,
		UnknownBuffer3:       []int{255, 255, 255, 255},
	}

	return playerData
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
