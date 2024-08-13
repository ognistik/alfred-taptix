*Forked from [Type-Joy](https://github.com/webdevcody/type-joy)*

* Fixed issue that modifiers were not recognized.
* Fixed mouse clicks not being initialized until key was pressed.
* Nocfree keyboard sound added.
* Mouse and Keyboard sounds are loaded separately, from different directories: `keyboards` and `mice` respectively, below the parent soundPath.
* Arguments now have their own flags, so they can be inserted in any order and without one another when running the app. Currently available flags are:

  * `-ks "Keyboard Sound"` for keyboard.
  * `-ms "Mouse Sound"` for mouse.
  * `-v 4.5` for global volume in scale 0 to 10 (Higher values are can be set, with possible distortion).
  * `-kv` and `-mv` for keyboard and mouse volumes, with number value from 0 to 10 (Higher values are can be set, with possible distortion).
  * `-sp "PATH"` can receive a different path for the parent folder of sounds.
  * If the `-mk` argument is passed, the app will initialize with the keyboard sounds muted. 
  * Similarly, `-mm` will mute the mouse sounds when initializing.

* The app can be controlled while running. Here’s the full list:

```
echo "set_keyboard newkeyboard" | nc localhost 8080
echo "set_mouse newmouse" | nc localhost 8080
echo "set_volume 4.5" | nc localhost 8080
echo "set_keyboard_volume 5" | nc localhost 8080
echo "set_mouse_volume 5" | nc localhost 8080
echo "toggle_keyboard" | nc localhost 8080
echo "toggle_mouse" | nc localhost 8080
echo "mute_keyboard" | nc localhost 8080
echo "mute_mouse" | nc localhost 8080
echo "unmute_keyboard" | nc localhost 8080
echo "unmute_mouse" | nc localhost 8080
echo "get_keyboard" | nc localhost 8080
echo "get_mouse" | nc localhost 8080
echo "get_volume" | nc localhost 8080
echo "get_keyboard_volume" | nc localhost 8080
echo "get_mouse_volume" | nc localhost 8080
echo "quit" | nc localhost 8080
```

Compile with 
```
go build -o type-joy-x main.go
```

Run in background once compiled with
```
nohup ./type-joy-x &
```

**All of these changes still need testing and, possibly, debugging… WIP**

---

This is a work in progress

# Virtualized Keyboard Switches

## How to Run

1. `go mod tidy`
2. `go run main.go`

## Planned Features

- add more keyboard sounds
- add more mouse sounds
- ability to customize click sounds

## Adding Sounds
