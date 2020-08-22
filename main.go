package main

import (
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/klyse/LogitechKeyboardLED/LogiKeyboard"
	"github.com/klyse/LogitechKeyboardLED/LogiKeyboardTypes"
	"github.com/moutend/go-hook/pkg/keyboard"
	"github.com/moutend/go-hook/pkg/types"
	"log"
	"os"
	"os/signal"
	"time"
)

var logiKeyboard *LogiKeyboard.LogiKeyboard

func main() {
	log.SetFlags(0)
	log.SetPrefix("error: ")

	logiKeyboard = LogiKeyboard.Create()

	defer logiKeyboard.Shutdown()

	logiKeyboard.Init()

	logiKeyboard.SetTargetDevice(LogiKeyboardTypes.LogiDeviceTypeAll)
	defaultLightning()

	var shortcuts = []Shortcut{
		*new(Shortcut).CreateWithKey([]types.VKCode{types.VK_LSHIFT}, []ShortcutKey{
			*new(ShortcutKey).CreateColor(LogiKeyboardTypes.F6, 100, 0, 0),
			*new(ShortcutKey).CreateColor(LogiKeyboardTypes.F9, 0, 0, 100),
		}),

		*new(Shortcut).CreateWithKey([]types.VKCode{types.VK_LCONTROL}, []ShortcutKey{
			*new(ShortcutKey).CreateColor(LogiKeyboardTypes.F2, 100, 0, 0),

			*new(ShortcutKey).CreateColor(LogiKeyboardTypes.C, 100, 0, 100),
			*new(ShortcutKey).CreateColor(LogiKeyboardTypes.V, 100, 0, 100),
			*new(ShortcutKey).CreateColor(LogiKeyboardTypes.X, 100, 0, 0),
			*new(ShortcutKey).CreateColor(LogiKeyboardTypes.Y, 100, 0, 0),

			*new(ShortcutKey).CreateColor(LogiKeyboardTypes.F, 0, 100, 0),
			*new(ShortcutKey).CreateColor(LogiKeyboardTypes.R, 100, 100, 0),
		}),

		*new(Shortcut).CreateWithKey([]types.VKCode{types.VK_LCONTROL, types.VK_LSHIFT}, []ShortcutKey{
			*new(ShortcutKey).CreateColor(LogiKeyboardTypes.T, 50, 0, 50),

			*new(ShortcutKey).CreateColor(LogiKeyboardTypes.F, 0, 100, 0),
			*new(ShortcutKey).CreateColor(LogiKeyboardTypes.R, 100, 100, 0),
		}),

		*new(Shortcut).CreateColor([]types.VKCode{types.VK_LWIN}, []LogiKeyboardTypes.Name{LogiKeyboardTypes.ONE, LogiKeyboardTypes.TWO, LogiKeyboardTypes.THREE, LogiKeyboardTypes.FOUR, LogiKeyboardTypes.FIVE, LogiKeyboardTypes.SIX, LogiKeyboardTypes.SEVEN, LogiKeyboardTypes.EIGHT, LogiKeyboardTypes.NINE, LogiKeyboardTypes.ZERO}, 0, 100, 0),
		*new(Shortcut).Create([]types.VKCode{types.VK_LMENU}, []LogiKeyboardTypes.Name{LogiKeyboardTypes.F4}),
	}

	if err := run(shortcuts); err != nil {
		log.Fatal(err)
	}
}

func defaultLightning() {
	fmt.Println("defaultLighting")
	logiKeyboard.SetLightning(100, 100, 100)
}

func run(shortcuts []Shortcut) error {
	// Buffer size is depends on your need. The 100 is placeholder value.
	keyboardChan := make(chan types.KeyboardEvent, 100)

	if err := keyboard.Install(nil, keyboardChan); err != nil {
		return err
	}

	defer keyboard.Uninstall()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	fmt.Println("start capturing keyboard input")

	currentlyPressedKeys := make(map[types.VKCode]bool)

	for {
		select {
		case <-time.After(5 * time.Minute):
			fmt.Println("Received timeout signal")
			return nil
		case <-signalChan:
			fmt.Println("Received shutdown signal")
			return nil
		case k := <-keyboardChan:
			fmt.Printf("Received %v %v\n", k.Message, k.VKCode)

			currentlyPressedKeys[k.VKCode] = k.Message == types.WM_KEYDOWN || k.Message == types.WM_SYSKEYDOWN

			shortCut := linq.From(shortcuts).Where(func(c interface{}) bool {
				found := linq.From(c.(Shortcut).modifiers).All(func(y interface{}) bool {
					return currentlyPressedKeys[y.(types.VKCode)]
				})

				if !found {
					return false
				}

				found = linq.From(currentlyPressedKeys).Where(func(y interface{}) bool {
					return y.(linq.KeyValue).Value.(bool)
				}).AnyWith(func(y interface{}) bool {
					return !linq.From(c.(Shortcut).modifiers).Contains(y.(linq.KeyValue).Key.(types.VKCode))
				})

				return !found
			}).First()

			if shortCut != nil {
				defaultLightning()
				for _, logiKey := range shortCut.(Shortcut).keys {
					logiKeyboard.SetLightingForKeyWithKeyName(logiKey.key, logiKey.red, logiKey.green, logiKey.blue)

				}
			} else {
				defaultLightning()
			}
		}

		continue
	}
}
