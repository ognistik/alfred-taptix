# Taptix Source Code
*Originally forked from [Type-Joy](https://github.com/webdevcody/type-joy)*

## Install Go
Taptix is an app built with Go.

Install Go with [Homebrew](https://brew.sh/):
`brew install go`

## How to Run
1. `go mod tidy`
2. `go run main.go`

## How to Compile
To compile for your current system:
`go build -o taptix main.go`

If you want to compile for both Apple Silicon and Intel, you may need to do something like this:
```
go mod tidy
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o taptix-arm64 main.go
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o taptix-amd64 main.go
lipo -create -output taptix taptix-arm64 taptix-amd64
```

Note... you may also need Xcode Command Line Tools installed: `xcode-select --install`

## Run the Compiled App as a Background Process
`nohup ./taptix &`

## Available Arguments and Flags
  You may use the following when running Taptix:
  * `-ks "Keyboard Sound"` for keyboard.
  * `-ms "Mouse Sound"` for mouse.
  * `-v 4.5` for global volume in scale 0 to 10 (Higher values can be set, with possible distortion).
  * `-kv` and `-mv` for keyboard and mouse volumes, with number values from 0 to 10 (Higher values can be set, with possible distortion).
  * `-kp "Keyboards Path"` and `-mp "Mice Path"` can each receive a different path for the parent folder of sounds.
  * If the `-mk` argument is passed, the app will initialize with the keyboard sounds muted. 
  * Similarly, `-mm` will mute the mouse sounds when initializing.

## Commands to Control Running Instance
The following can control Taptix while running, even if it's running in the background.

```bash
echo "set_keyboard \"new keyboard\" \"keyboards path\"" | nc localhost 8080 # The Keyboard path is optional. Sounds should still be inside a "keyboards" folder.
echo "set_mouse \"new mouse\" \"mice path\"" | nc localhost 8080 # The Mice path is optional. Sounds should still be inside a "mice" folder.
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