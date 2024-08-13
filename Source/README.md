# Taptix Source Code
*Originally forked from [Type-Joy](https://github.com/webdevcody/type-joy)*

## Getting Go
Taptix is an app built with Go.

Install Go with [Homebrew](https://brew.sh/):
`brew install go`

## How to Run
1. `go mod tidy`
2. `go run main.go`

## How to Compile
`go build -o type-joy main.go`

## Run the Compiled App as a Background Process
`nohup ./type-joy &`

## Available Arguments and Flags
  You may use the following when running Taptix:
  * `-ks "Keyboard Sound"` for keyboard.
  * `-ms "Mouse Sound"` for mouse.
  * `-v 4.5` for global volume in scale 0 to 10 (Higher values are can be set, with possible distortion).
  * `-kv` and `-mv` for keyboard and mouse volumes, with number value from 0 to 10 (Higher values are can be set, with possible distortion).
  * `-sp "PATH"` can receive a different path for the parent folder of sounds.
  * If the `-mk` argument is passed, the app will initialize with the keyboard sounds muted. 
  * Similarly, `-mm` will mute the mouse sounds when initializing.

## Commands to Control Running Instance
The following can control Taptix while running, even if it's running in the background.

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

## Additions and Fixes from Original Code
* Fixed issue where modifiers were not recognized.
* Fixed mouse clicks not being initialized until key was pressed.
* Mouse and Keyboard sounds are loaded separately, from different directories: `keyboards` and `mice` respectively, below the parent soundPath.
* Added separate global and per device volume and muting controls.
* Implemented arguments to be used with their own flags. Arguments can be inserted in any order and without one another when running the app.
* Added a "listener" and commands to control Taptix while running.