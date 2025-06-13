package game

import (
	"sort"
	"strings"
)

type Item struct {
	Name     string
	Wearable bool
	Usable   bool
}

type Room struct {
	Name        string
	Description string
	Items       map[string]*Item
	Connections []string
	OnEnter     func() string
	OnLook      func() string
}

type Player struct {
	CurrentRoom *Room
	Inventory   map[string]*Item
	HasBackpack bool
}

type GameState struct {
	Player    *Player
	Rooms     map[string]*Room
	DoorState map[string]bool
}

var gameState *GameState

func initGame() {
	keys := &Item{Name: "ключи", Usable: true}
	notes := &Item{Name: "конспекты"}
	backpack := &Item{Name: "рюкзак", Wearable: true}
	tea := &Item{Name: "чай"}

	kitchen := &Room{
		Name:        "кухня",
		Description: "кухня, ничего интересного.",
		Items:       map[string]*Item{"чай": tea},
		Connections: []string{"коридор"},
		OnEnter: func() string {
			return "кухня, ничего интересного. можно пройти - коридор"
		},
		OnLook: func() string {
			if gameState.Player.HasBackpack {
				return "ты находишься на кухне, на столе: чай, надо идти в универ. можно пройти - коридор"
			}
			return "ты находишься на кухне, на столе: чай, надо собрать рюкзак и идти в универ. можно пройти - коридор"
		},
	}

	corridor := &Room{
		Name:        "коридор",
		Description: "ничего интересного.",
		Connections: []string{"кухня", "комната", "улица"},
	}

	room := &Room{
		Name: "комната",
		OnEnter: func() string {
			return "ты в своей комнате. можно пройти - коридор"
		},
		OnLook: func() string {
			var items []string
			for name := range gameState.Rooms["комната"].Items {
				items = append(items, name)
			}
			sort.Strings(items)

			var tableItems, chairItems []string
			for _, name := range items {
				item := gameState.Rooms["комната"].Items[name]
				if item.Name == "рюкзак" {
					chairItems = append(chairItems, item.Name)
				} else {
					tableItems = append(tableItems, item.Name)
				}
			}

			msg := ""
			if len(tableItems) > 0 {
				msg += "на столе: " + strings.Join(tableItems, ", ")
			}
			if len(chairItems) > 0 {
				if msg != "" {
					msg += ", "
				}
				msg += "на стуле: " + strings.Join(chairItems, ", ")
			}

			if msg == "" {
				return "пустая комната. можно пройти - коридор"
			}
			return msg + ". можно пройти - коридор"
		},
		Items: map[string]*Item{
			"ключи":     keys,
			"конспекты": notes,
			"рюкзак":    backpack,
		},
		Connections: []string{"коридор"},
	}

	outside := &Room{
		Name:        "улица",
		Description: "на улице весна.",
		Connections: []string{"коридор"},
		OnLook: func() string {
			return "на улице весна. можно пройти - домой"
		},
	}

	gameState = &GameState{
		Player: &Player{
			CurrentRoom: kitchen,
			Inventory:   make(map[string]*Item),
		},
		Rooms: map[string]*Room{
			"кухня":   kitchen,
			"коридор": corridor,
			"комната": room,
			"улица":   outside,
		},
		DoorState: map[string]bool{
			"улица": true,
		},
	}
}
func handleCommand(command string) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "неизвестная команда"
	}

	cmd := parts[0]
	switch cmd {
	case "осмотреться":
		return lookAround()
	case "идти":
		if len(parts) < 2 {
			return "неизвестная команда"
		}
		return moveTo(parts[1])
	case "надеть":
		if len(parts) < 2 {
			return "неизвестная команда"
		}
		return wear(parts[1])
	case "взять":
		if len(parts) < 2 {
			return "неизвестная команда"
		}
		return takeItem(parts[1])
	case "применить":
		if len(parts) < 3 {
			return "неизвестная команда"
		}
		return useItem(parts[1], parts[2])
	default:
		return "неизвестная команда"
	}
}

func lookAround() string {
	if gameState.Player.CurrentRoom.OnLook != nil {
		return gameState.Player.CurrentRoom.OnLook()
	}
	return gameState.Player.CurrentRoom.Description + " можно пройти - " +
		strings.Join(gameState.Player.CurrentRoom.Connections, ", ")
}

func moveTo(roomName string) string {
	current := gameState.Player.CurrentRoom

	for _, conn := range current.Connections {
		if conn == roomName {
			if roomName == "улица" && gameState.DoorState["улица"] {
				return "дверь закрыта"
			}

			gameState.Player.CurrentRoom = gameState.Rooms[roomName]

			if gameState.Player.CurrentRoom.OnEnter != nil {
				return gameState.Player.CurrentRoom.OnEnter()
			}
			return lookAround()
		}
	}
	return "нет пути в " + roomName
}

func wear(itemName string) string {
	if gameState.Player.CurrentRoom.Name != "комната" {
		return "неизвестная команда"
	}

	item, exists := gameState.Player.CurrentRoom.Items[itemName]
	if !exists || !item.Wearable {
		return "неизвестная команда"
	}

	if gameState.Player.HasBackpack {
		return "вы уже надели: рюкзак"
	}

	gameState.Player.HasBackpack = true
	delete(gameState.Player.CurrentRoom.Items, itemName)
	return "вы надели: рюкзак"
}

func takeItem(itemName string) string {
	if !gameState.Player.HasBackpack {
		return "некуда класть"
	}

	item, exists := gameState.Player.CurrentRoom.Items[itemName]
	if !exists {
		return "нет такого"
	}

	gameState.Player.Inventory[itemName] = item
	delete(gameState.Player.CurrentRoom.Items, itemName)
	return "предмет добавлен в инвентарь: " + itemName
}

func useItem(itemName, target string) string {
	_, exists := gameState.Player.Inventory[itemName]
	if !exists {
		return "нет предмета в инвентаре - " + itemName
	}

	if target == "дверь" && itemName == "ключи" {
		gameState.DoorState["улица"] = false
		return "дверь открыта"
	}

	return "не к чему применить"
}
