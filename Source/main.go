package main

import (
    "bytes"
    "flag"
    "fmt"
    "io"
    "log"
    "math/rand"
    "net"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "sync"

    "github.com/hajimehoshi/go-mp3"
    "github.com/hajimehoshi/oto"
    hook "github.com/robotn/gohook"
)

var (
    keyboard       string
    mouse          string
    volume         float64
    keyboardVolume float64
    mouseVolume    float64
    keyboardPath   string
    mousePath      string
    muteKeyboard   bool
    muteMouse      bool
	wg sync.WaitGroup
    context *oto.Context
	isMouseMuted   bool
    mouseMutex     sync.Mutex
)

const (
    ENTER uint16 = 36
    SPACE uint16 = 49
)

var sounds map[string][]byte = make(map[string][]byte)
var keyMap = make(map[uint16]bool)

func setKeyboard(conn net.Conn, newKeyboard, newPath string) {
    path := keyboardPath
    if newPath != "" {
        path = newPath
    }
    if _, err := os.Stat(filepath.Join(path, "keyboards", newKeyboard)); os.IsNotExist(err) {
        conn.Write([]byte(fmt.Sprintf("Keyboard sounds not found: %s\n", newKeyboard)))
    } else {
        keyboard = newKeyboard
        if newPath != "" {
            keyboardPath = newPath
        }
        loadSoundsForKeyboard(keyboard)
        conn.Write([]byte(fmt.Sprintf("Keyboard set to: %s\n", keyboard)))
    }
}

func setMouse(conn net.Conn, newMouse, newPath string) {
    path := mousePath
    if newPath != "" {
        path = newPath
    }
    if _, err := os.Stat(filepath.Join(path, "mice", newMouse)); os.IsNotExist(err) {
        conn.Write([]byte(fmt.Sprintf("Mouse sounds not found: %s\n", newMouse)))
    } else {
        mouse = newMouse
        if newPath != "" {
            mousePath = newPath
        }
        loadSoundsForMouse(mouse)
        conn.Write([]byte(fmt.Sprintf("Mouse set to: %s\n", mouse)))
    }
}

func loadSoundsForKeyboard(keyboard string) {
    keys := []string{"down1", "up1", "down2", "up2", "down3", "up3", "down4", "up4", "down5", "up5", "down6", "up6", "down7", "up7", "down_space", "up_space", "down_enter", "up_enter"}
    for _, key := range keys {
        loadSound("keyboard", keyboard, key)
    }
}

func loadSoundsForMouse(mouse string) {
    keys := []string{"up_mouse", "down_mouse"}
    for _, key := range keys {
        loadSound("mouse", mouse, key)
    }
}

func loadSound(deviceType, device, soundName string) {
    var soundFile *os.File
    var err error

    if deviceType == "keyboard" {
        soundFile, err = os.Open(filepath.Join(keyboardPath, "keyboards", device, soundName+".mp3"))
    } else {
        soundFile, err = os.Open(filepath.Join(mousePath, "mice", device, soundName+".mp3"))
    }

    if err != nil {
        log.Fatalf("failed to open sound file: %v", err)
    }
    defer soundFile.Close()

    sound, err := io.ReadAll(soundFile)
    if err != nil {
        log.Fatalf("failed to read sound file: %v", err)
    }

    sounds[soundName] = sound
}

func getRandomDownKey() string {
    keys := []string{"down1", "down2", "down3", "down4", "down5", "down6", "down7"}
    return keys[rand.Intn(len(keys))]
}

func getRandomUpKey() string {
    keys := []string{"up1", "up2", "up3", "up4", "up5", "up6", "up7"}
    return keys[rand.Intn(len(keys))]
}

func adjustVolume(samples []byte, volume float64) []byte {
    for i := 0; i < len(samples); i += 2 {
        sample := int16(int(samples[i]) | int(samples[i+1])<<8)
        sample = int16(float64(sample) * volume)
        samples[i] = byte(sample)
        samples[i+1] = byte(sample >> 8)
    }
    return samples
}

func playSound(key string, deviceVolume float64) {
    decoder, err := mp3.NewDecoder(bytes.NewReader(sounds[key]))
    if err != nil {
        log.Fatalf("failed to create MP3 decoder: %v", err)
    }

    player := context.NewPlayer()
    defer player.Close()

    decoder.Seek(0, 0)

    buf := make([]byte, 8192)
    for {
        n, err := decoder.Read(buf)
        if err != nil && err != io.EOF {
            log.Printf("failed to read decoded audio: %v", err)
            break
        }
        if n == 0 {
            break
        }
        adjustedBuf := adjustVolume(buf[:n], volume*deviceVolume)
        player.Write(adjustedBuf)
    }
}

func scheduleSound(key string, deviceVolume float64) {
    // useful to help debug sounds
    // fmt.Printf("Playing sound: %s\n", key)
    wg.Add(1)
    go func() {
        defer wg.Done()
        playSound(key, deviceVolume)
    }()
}

func registerMouseHandlers() {
    hook.Register(hook.MouseHold, []string{}, func(e hook.Event) {
        mouseMutex.Lock()
        if isMouseMuted {
            mouseMutex.Unlock()
            return
        }
        mouseMutex.Unlock()

        if keyMap[e.Button] {
            return
        }
        keyMap[e.Button] = true
        scheduleSound("down_mouse", mouseVolume)
    })

    hook.Register(hook.MouseDown, []string{}, func(e hook.Event) {
        mouseMutex.Lock()
        if isMouseMuted {
            mouseMutex.Unlock()
            return
        }
        mouseMutex.Unlock()

        if keyMap[e.Button] {
            scheduleSound("up_mouse", mouseVolume)
            keyMap[e.Button] = false
            return
        }
    })

    hook.Register(hook.MouseUp, []string{}, func(e hook.Event) {
        mouseMutex.Lock()
        if isMouseMuted {
            mouseMutex.Unlock()
            return
        }
        mouseMutex.Unlock()

        if !keyMap[e.Button] {
            return
        }
        keyMap[e.Button] = false
        scheduleSound("up_mouse", mouseVolume)
    })
}

func registerMutedMouseHandlers() {
    hook.Register(hook.MouseHold, []string{}, func(e hook.Event) {})
    hook.Register(hook.MouseDown, []string{}, func(e hook.Event) {})
    hook.Register(hook.MouseUp, []string{}, func(e hook.Event) {})
}

func registerUnmutedMouseHandlers() {
    hook.Register(hook.MouseHold, []string{}, func(e hook.Event) {
        if keyMap[e.Button] {
            return
        }
        keyMap[e.Button] = true
        scheduleSound("down_mouse", mouseVolume)
    })

    hook.Register(hook.MouseDown, []string{}, func(e hook.Event) {
        if keyMap[e.Button] {
            scheduleSound("down_mouse", mouseVolume)
            keyMap[e.Button] = false
            return
        }
    })

    hook.Register(hook.MouseUp, []string{}, func(e hook.Event) {
        if !keyMap[e.Button] {
            return
        }
        keyMap[e.Button] = false
        scheduleSound("up_mouse", mouseVolume)
    })
}

func parseQuotedCommand(command string) []string {
    var parts []string
    var part strings.Builder
    inQuotes := false

    for _, r := range command {
        switch r {
        case '"':
            inQuotes = !inQuotes
        case ' ':
            if !inQuotes {
                if part.Len() > 0 {
                    parts = append(parts, part.String())
                    part.Reset()
                }
            } else {
                part.WriteRune(r)
            }
        default:
            part.WriteRune(r)
        }
    }

    if part.Len() > 0 {
        parts = append(parts, part.String())
    }

    return parts
}

func main() {
    flag.StringVar(&keyboard, "ks", "Nocfree Lite", "Keyboard sound")
    flag.StringVar(&mouse, "ms", "Magic Mouse", "Mouse sound")
    flag.Float64Var(&volume, "v", 10.0, "Global volume (0-10)")
    flag.Float64Var(&keyboardVolume, "kv", 10.0, "Keyboard volume (0-10)")
    flag.Float64Var(&mouseVolume, "mv", 10.0, "Mouse volume (0-10)")
    flag.StringVar(&keyboardPath, "kp", "sounds", "Keyboard sounds path")
    flag.StringVar(&mousePath, "mp", "sounds", "Mouse sounds path")
    flag.BoolVar(&muteKeyboard, "mk", false, "Mute keyboard sounds")
    flag.BoolVar(&muteMouse, "mm", false, "Mute mouse sounds")
    flag.Parse()

    // Convert volume from 0-10 scale to 0-1 scale
    volume = volume / 10
    keyboardVolume = keyboardVolume / 10
    mouseVolume = mouseVolume / 10

    if _, err := os.Stat(filepath.Join(keyboardPath, "keyboards", keyboard)); os.IsNotExist(err) {
        log.Fatalf("Keyboard sounds not found: %s", keyboard)
    }

    if _, err := os.Stat(filepath.Join(mousePath, "mice", mouse)); os.IsNotExist(err) {
        log.Fatalf("Mouse sounds not found: %s", mouse)
    }

    fmt.Printf("Using keyboard: %s\n", keyboard)
    fmt.Printf("Using mouse: %s\n", mouse)
    fmt.Printf("Global volume: %.2f\n", volume*10)
    if muteKeyboard {
        fmt.Println("Keyboard volume: muted")
    } else {
        fmt.Printf("Keyboard volume: %.2f\n", keyboardVolume*10)
    }
    
    if muteMouse {
        fmt.Println("Mouse volume: muted")
    } else {
        fmt.Printf("Mouse volume: %.2f\n", mouseVolume*10)
    }

    loadSoundsForKeyboard(keyboard)
    loadSoundsForMouse(mouse)

    var err error
    context, err = oto.NewContext(48000, 2, 2, 8192)
    if err != nil {
        log.Fatalf("failed to create Oto context: %v", err)
    }
    defer context.Close()

    hook.Register(hook.KeyHold, []string{}, func(e hook.Event) {
        if keyMap[e.Rawcode] || muteKeyboard {
            return
        }
        keyMap[e.Rawcode] = true

        if e.Rawcode == ENTER {
            scheduleSound("down_enter", keyboardVolume)
        } else if e.Rawcode == SPACE {
            scheduleSound("down_space", keyboardVolume)
        } else {
            scheduleSound(getRandomDownKey(), keyboardVolume)
        }
    })

    hook.Register(hook.KeyDown, []string{"A-Z a-z 0-9"}, func(e hook.Event) {
        if keyMap[e.Rawcode] || muteKeyboard {
            return
        }
        keyMap[e.Rawcode] = true

        if e.Rawcode == ENTER {
            scheduleSound("down_enter", keyboardVolume)
        } else if e.Rawcode == SPACE {
            scheduleSound("down_space", keyboardVolume)
        } else {
            scheduleSound(getRandomDownKey(), keyboardVolume)
        }
    })

    hook.Register(hook.KeyUp, []string{"A-Z a-z 0-9"}, func(e hook.Event) {
        if !keyMap[e.Rawcode] || muteKeyboard {
            return
        }
        keyMap[e.Rawcode] = false

        if e.Rawcode == ENTER {
            scheduleSound("up_enter", keyboardVolume)
        } else if e.Rawcode == SPACE {
            scheduleSound("up_space", keyboardVolume)
        } else {
            scheduleSound(getRandomUpKey(), keyboardVolume)
        }
    })

	isMouseMuted = muteMouse
    registerMouseHandlers()

    go func() {
        listener, err := net.Listen("tcp", "localhost:8080")
        if err != nil {
            log.Fatal(err)
        }
        defer listener.Close()

        for {
            conn, err := listener.Accept()
            if err != nil {
                log.Println(err)
                continue
            }
            go handleConnection(conn)
        }
    }()

    s := hook.Start()
    <-hook.Process(s)
    wg.Wait()
}

func handleConnection(conn net.Conn) {
    defer conn.Close()

    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        log.Println(err)
        return
    }

    command := strings.TrimSpace(string(buffer[:n]))
    parts := parseQuotedCommand(command)

    switch parts[0] {
    case "set_keyboard":
        if len(parts) < 2 || len(parts) > 3 {
            conn.Write([]byte("Invalid command. Usage: set_keyboard \"<new_keyboard>\" [\"<new_path>\"]\n"))
            return
        }
        newKeyboard := parts[1]
        newPath := ""
        if len(parts) == 3 {
            newPath = parts[2]
        }
        setKeyboard(conn, newKeyboard, newPath)
    case "set_mouse":
        if len(parts) < 2 || len(parts) > 3 {
            conn.Write([]byte("Invalid command. Usage: set_mouse \"<new_mouse>\" [\"<new_path>\"]\n"))
            return
        }
        newMouse := parts[1]
        newPath := ""
        if len(parts) == 3 {
            newPath = parts[2]
        }
        setMouse(conn, newMouse, newPath)
    case "set_volume":
        if len(parts) > 1 {
            newVolume, err := parseVolume(parts[1])
            if err == nil {
                volume = newVolume
                conn.Write([]byte(fmt.Sprintf("Global volume set to: %.2f\n", volume*10)))
            } else {
                conn.Write([]byte("Invalid volume value\n"))
            }
        }
    case "set_keyboard_volume":
        if len(parts) > 1 {
            newVolume, err := parseVolume(parts[1])
            if err == nil {
                keyboardVolume = newVolume
                conn.Write([]byte(fmt.Sprintf("Keyboard volume set to: %.2f\n", keyboardVolume*10)))
            } else {
                conn.Write([]byte("Invalid volume value\n"))
            }
        }
    case "set_mouse_volume":
        if len(parts) > 1 {
            newVolume, err := parseVolume(parts[1])
            if err == nil {
                mouseVolume = newVolume
                conn.Write([]byte(fmt.Sprintf("Mouse volume set to: %.2f\n", mouseVolume*10)))
            } else {
                conn.Write([]byte("Invalid volume value\n"))
            }
        }
    case "toggle_keyboard":
        muteKeyboard = !muteKeyboard
        if muteKeyboard {
            conn.Write([]byte("Keyboard sounds muted\n"))
        } else {
            conn.Write([]byte("Keyboard sounds unmuted\n"))
        }
    case "mute_keyboard":
        muteKeyboard = true
        conn.Write([]byte("Keyboard sounds muted\n"))
    case "unmute_keyboard":
        muteKeyboard = false
        conn.Write([]byte("Keyboard sounds unmuted\n"))
    case "toggle_mouse":
		mouseMutex.Lock()
		muteMouse = !muteMouse
		isMouseMuted = muteMouse
		mouseMutex.Unlock()
		if muteMouse {
			conn.Write([]byte("Mouse sounds muted\n"))
		} else {
			conn.Write([]byte("Mouse sounds unmuted\n"))
		}
    case "mute_mouse":
        mouseMutex.Lock()
        muteMouse = true
        isMouseMuted = muteMouse
        mouseMutex.Unlock()
        conn.Write([]byte("Mouse sounds muted\n"))
    case "unmute_mouse":
        mouseMutex.Lock()
        muteMouse = false
        isMouseMuted = muteMouse
        mouseMutex.Unlock()
        conn.Write([]byte("Mouse sounds unmuted\n"))
    case "get_keyboard":
        conn.Write([]byte(fmt.Sprintf("Current keyboard: %s\n", keyboard)))
    case "get_mouse":
        conn.Write([]byte(fmt.Sprintf("Current mouse: %s\n", mouse)))
    case "get_volume":
        conn.Write([]byte(fmt.Sprintf("Current global volume: %.2f\n", volume*10)))
    case "get_keyboard_volume":
        if muteKeyboard {
            conn.Write([]byte("Keyboard is muted\n"))
        } else {
            conn.Write([]byte(fmt.Sprintf("Current keyboard volume: %.2f\n", keyboardVolume*10)))
        }
    case "get_mouse_volume":
        if muteMouse {
            conn.Write([]byte("Mouse is muted\n"))
        } else {
            conn.Write([]byte(fmt.Sprintf("Current mouse volume: %.2f\n", mouseVolume*10)))
        }
    case "quit":
        conn.Write([]byte("Shutting down...\n"))
        os.Exit(0)
    default:
        conn.Write([]byte("Unknown command\n"))
    }
}

func parseVolume(volumeStr string) (float64, error) {
    volume, err := strconv.ParseFloat(volumeStr, 64)
    if err != nil {
        return 0, err
    }
    if volume < 0 || volume > 10 {
        return 0, fmt.Errorf("volume must be between 0 and 10")
    }
    return volume / 10, nil
}