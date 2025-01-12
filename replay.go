package polytopiamapmodel

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

type ActionBuild struct {
	PlayerId        uint8
	ImprovementType uint16
	Coordinates     [2]uint32
}

type ActionAttack struct {
	PlayerId uint8
	UnitId   uint32
	Origin   [2]uint32
	Target   [2]uint32
}

type ActionRecover struct {
	PlayerId    uint8
	Coordinates [2]uint32
}

type ActionTrain struct {
	PlayerId uint8
	UnitType uint16
	Position [2]uint32
}

type ActionMove struct {
	PlayerId    uint8
	OldPosition [2]uint32
	NewPosition [2]uint32
	UnitId      uint32
}

type ActionCaptureCity struct {
	PlayerId    uint8
	UnitId      uint32
	Coordinates [2]uint32
}

type ActionResearch struct {
	PlayerId uint8
	TechType uint16
}

type ActionDestroyImprovement struct {
	PlayerId    uint8
	Coordinates [2]uint32
}

type ActionCityReward struct {
	PlayerId    uint8
	Coordinates [2]uint32
	Reward      uint16
}

type ActionPromote struct {
	PlayerId    uint8
	Coordinates [2]uint32
}

type ActionExamineRuins struct {
	PlayerId    uint8
	Coordinates [2]uint32
}

type ActionEndTurn struct {
	PlayerId uint8
}

type ActionUpgrade struct {
	PlayerId    uint8
	UnitType    uint16
	Coordinates [2]uint32
}

type ActionCityLevelUp struct {
	PlayerId    uint8
	Coordinates [2]uint32
}

func readAllActions(streamReader *io.SectionReader) map[int][]ActionCaptureCity {
	numActions := unsafeReadUint16(streamReader)

	turnCaptureMap := make(map[int][]ActionCaptureCity)

	replayActions := make([]string, 0)
	turn := 1
	for i := 0; i < int(numActions); i++ {
		actionType := unsafeReadUint16(streamReader)

		var buffer []byte
		if actionType == 1 {
			action := ActionBuild{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
			replayActions = append(replayActions, fmt.Sprintf("Build: %+v\n", action))
		} else if actionType == 2 {
			action := ActionAttack{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
			replayActions = append(replayActions, fmt.Sprintf("Attack: %+v\n", action))
		} else if actionType == 3 {
			action := ActionRecover{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
			replayActions = append(replayActions, fmt.Sprintf("Recover: %+v\n", action))
		} else if actionType == 4 {
			buffer = readFixedList(streamReader, 9)
		} else if actionType == 5 {
			action := ActionTrain{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
			replayActions = append(replayActions, fmt.Sprintf("Train: %+v\n", action))
		} else if actionType == 6 {
			action := ActionMove{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
			replayActions = append(replayActions, fmt.Sprintf("Move: %+v\n", action))
		} else if actionType == 7 {
			action := ActionCaptureCity{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
			replayActions = append(replayActions, fmt.Sprintf("CaptureCity: %+v\n", action))

			_, ok := turnCaptureMap[turn]
			if !ok {
				turnCaptureMap[turn] = make([]ActionCaptureCity, 0)
			}
			turnCaptureMap[turn] = append(turnCaptureMap[turn], action)
		} else if actionType == 8 {
			action := ActionResearch{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
			replayActions = append(replayActions, fmt.Sprintf("Research: %+v\n", action))
		} else if actionType == 9 {
			action := ActionDestroyImprovement{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
			replayActions = append(replayActions, fmt.Sprintf("DestroyImprovement: %+v\n", action))
		} else if actionType == 11 {
			action := ActionCityReward{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
			replayActions = append(replayActions, fmt.Sprintf("CityReward: %+v\n", action))
		} else if actionType == 13 {
			action := ActionPromote{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
			replayActions = append(replayActions, fmt.Sprintf("Promote: %+v\n", action))
		} else if actionType == 14 {
			action := ActionExamineRuins{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
			replayActions = append(replayActions, fmt.Sprintf("ExamineRuins: %+v\n", action))
		} else if actionType == 15 {
			action := ActionEndTurn{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
			replayActions = append(replayActions, fmt.Sprintf("EndTurn: %+v\n", action))

			if action.PlayerId == 255 {
				turn++
				replayActions = append(replayActions, fmt.Sprintf("Start new turn %v", turn))
			}
		} else if actionType == 16 {
			action := ActionUpgrade{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
			replayActions = append(replayActions, fmt.Sprintf("Upgrade: %+v\n", action))
		} else if actionType == 17 {
			buffer = readFixedList(streamReader, 9)
		} else if actionType == 18 {
			buffer = readFixedList(streamReader, 9)
		} else if actionType == 20 {
			buffer = readFixedList(streamReader, 1)
		} else if actionType == 21 {
			action := ActionCityLevelUp{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
			replayActions = append(replayActions, fmt.Sprintf("CityLevelUp: %+v\n", action))
		} else if actionType == 24 {
			buffer = readFixedList(streamReader, 9)
		} else if actionType == 25 {
			buffer = readFixedList(streamReader, 9)
		} else if actionType == 27 {
			buffer = readFixedList(streamReader, 10)
		} else if actionType == 28 {
			buffer = readFixedList(streamReader, 3)
		} else if actionType == 29 {
			buffer = readFixedList(streamReader, 10)
		} else if actionType == 30 {
			buffer = readFixedList(streamReader, 10)
		} else {
			log.Fatal("Unknown action type:", actionType)
		}

		if len(buffer) > 0 {
			replayActions = append(replayActions, fmt.Sprintf("Index", i, ", action type:", actionType, ", buffer:", buffer))
		}
	}

	return turnCaptureMap
}
