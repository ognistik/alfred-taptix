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
    "time"

    "github.com/hajimehoshi/go-mp3"
    "github.com/hajimehoshi/oto"
    hook "github.com/robotn/gohook"
    "github.com/youpy/go-wav"
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
    wg             sync.WaitGroup
    context        *oto.Context
    contextMutex   sync.Mutex
    isMouseMuted   bool
    mouseMutex     sync.Mutex
    once           sync.Once
)

const (
    ENTER           uint16 = 36
    SPACE           uint16 = 49
    THRESHOLD_MS    int64  = 10 // Threshold for key/mouse up events in milliseconds
    MULTI_KEY_DELAY int64  = 20 // Delay for multi-key presses in milliseconds
)

var sounds map[string]*WavData = make(map[string]*WavData)
var keyMap = make(map[uint16]bool)
var lastKeyPressTime = make(map[uint16]int64)
var lastKeyDownTime int64
var lastKeyUpTime int64
var keyState = make(map[uint16]bool)
var lastMouseDownTime = make(map[uint16]int64)
var lastMouseUpTime = make(map[uint16]int64)

func initContext() {
    once.Do(func() {
        var err error
        // Create a context with maximum quality settings
        context, err = oto.NewContext(48000, 2, 2, 8192)
        if err != nil {
            log.Fatalf("failed to create oto context: %v", err)
        }
    })
}

func convert24To16(data []byte) []byte {
    output := make([]byte, len(data)/3*2)
    for i := 0; i < len(data); i += 3 {
        // Convert 24-bit to 16-bit by dropping the least significant byte
        output[i/3*2] = data[i+1]
        output[i/3*2+1] = data[i+2]
    }
    return output
}

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
    // Clear existing keyboard sounds
    clearKeyboardSounds()
    
    keys := []string{"down1", "up1", "down2", "up2", "down3", "up3", "down4", "up4", "down5", "up5", "down6", "up6", "down7", "up7", "down_space", "up_space", "down_enter", "up_enter"}
    for _, key := range keys {
        loadSound("keyboard", keyboard, key)
    }
}

func loadSoundsForMouse(mouse string) {
    // Clear existing mouse sounds
    clearMouseSounds()

    keys := []string{"up_mouse", "down_mouse"}
    for _, key := range keys {
        loadSound("mouse", mouse, key)
    }
}

func clearKeyboardSounds() {
    keyboardKeys := []string{"down1", "up1", "down2", "up2", "down3", "up3", "down4", "up4", "down5", "up5", "down6", "up6", "down7", "up7", "down_space", "up_space", "down_enter", "up_enter"}
    for _, key := range keyboardKeys {
        delete(sounds, key+".wav")
        delete(sounds, key+".mp3")
    }
}

func clearMouseSounds() {
    mouseKeys := []string{"up_mouse", "down_mouse"}
    for _, key := range mouseKeys {
        delete(sounds, key+".wav")
        delete(sounds, key+".mp3")
    }
}

// loadSound attempts to load a WAV file first, then falls back to MP3 if WAV is not found
func loadSound(deviceType, device, soundName string) {
    var soundPath string
    if deviceType == "keyboard" {
        soundPath = filepath.Join(keyboardPath, "keyboards", device)
    } else {
        soundPath = filepath.Join(mousePath, "mice", device)
    }

    wavPath := filepath.Join(soundPath, soundName+".wav")
    mp3Path := filepath.Join(soundPath, soundName+".mp3")

    // Try loading WAV file first
    if wavData, err := loadWavFile(wavPath); err == nil {
        sounds[soundName+".wav"] = wavData
        log.Printf("Loaded WAV file: %s", soundName+".wav")
        return
    }

    // If WAV file not found, try loading MP3 file
    if data, err := os.ReadFile(mp3Path); err == nil {
        sounds[soundName+".mp3"] = &WavData{Data: data, Format: nil}
        log.Printf("Loaded MP3 file: %s", soundName+".mp3")
        return
    }

    log.Fatalf("failed to load sound file for %s", soundName)
}

type WavData struct {
    Data   []byte
    Format *wav.WavFormat
}

func loadWavFile(path string) (*WavData, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    reader := wav.NewReader(file)
    format, err := reader.Format()
    if err != nil {
        return nil, err
    }

    data, err := io.ReadAll(reader)
    if err != nil {
        return nil, err
    }

    return &WavData{Data: data, Format: format}, nil
}

func loadMp3File(path string) ([]byte, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    return io.ReadAll(file)
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

func playSound(soundName string, deviceVolume float64) {
    initContext() // This will only initialize the context once

    wavKey := soundName + ".wav"
    mp3Key := soundName + ".mp3"

    var wavData *WavData
    var isWav bool

    if data, ok := sounds[wavKey]; ok {
        wavData = data
        isWav = true
    } else if data, ok := sounds[mp3Key]; ok {
        wavData = data
        isWav = false
    } else {
        log.Printf("Sound not found for: %s", soundName)
        return
    }

    var stream io.Reader
    var err error

    if isWav {
        format := wavData.Format
        if format.BitsPerSample == 24 {
            // Convert 24-bit to 16-bit
            wavData.Data = convert24To16(wavData.Data)
            format.BitsPerSample = 16
        }

        // Resample if necessary
        if format.SampleRate != 48000 {
            wavData.Data = resampleAudio(wavData.Data, int(format.SampleRate), 48000, int(format.NumChannels))
        }

        stream = bytes.NewReader(wavData.Data)
    } else {
        stream, err = mp3.NewDecoder(bytes.NewReader(wavData.Data))
        if err != nil {
            log.Printf("failed to create MP3 decoder for %s: %v", soundName, err)
            return
        }
    }

    contextMutex.Lock()
    player := context.NewPlayer()
    contextMutex.Unlock()
    defer player.Close()

    // Read and play the audio data in chunks
    buffer := make([]byte, 4096)
    for {
        n, err := stream.Read(buffer)
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Printf("error reading audio data for %s: %v", soundName, err)
            return
        }

        adjustedData := adjustVolume(buffer[:n], volume*deviceVolume)
        _, err = player.Write(adjustedData)
        if err != nil {
            log.Printf("failed to write audio data for %s: %v", soundName, err)
            return
        }
    }

    // Ensure all audio data is played
    player.Close()
}

func resampleAudio(data []byte, fromSampleRate, toSampleRate, channels int) []byte {
    // Implement a simple linear interpolation resampling
    // This is a basic implementation and might not provide the best audio quality
    // For better quality, consider using a library like https://github.com/go-audio/audio
    ratio := float64(toSampleRate) / float64(fromSampleRate)
    outputLength := int(float64(len(data)) * ratio)
    output := make([]byte, outputLength)

    for i := 0; i < outputLength; i += 2 * channels {
        position := float64(i) / ratio
        index := int(position)
        if index >= len(data)-2*channels {
            break
        }
        for c := 0; c < channels; c++ {
            low := int16(data[index+c*2]) | int16(data[index+c*2+1])<<8
            high := int16(data[index+c*2+2*channels]) | int16(data[index+c*2+1+2*channels])<<8
            value := low + int16(float64(high-low)*(position-float64(index)))
            output[i+c*2] = byte(value)
            output[i+c*2+1] = byte(value >> 8)
        }
    }

    return output
}

func scheduleSound(key string, deviceVolume float64) {
    wg.Add(1)
    go func() {
        defer wg.Done()
        playSound(key, deviceVolume)
    }()
}

func registerKeyboardHandlers() {
    hook.Register(hook.KeyHold, []string{}, func(e hook.Event) {
        handleKeyEvent(e, true)
    })

    hook.Register(hook.KeyDown, []string{"A-Z a-z 0-9"}, func(e hook.Event) {
        handleKeyEvent(e, true)
    })

    hook.Register(hook.KeyUp, []string{"A-Z a-z 0-9"}, func(e hook.Event) {
        handleKeyEvent(e, false)
    })
}

func handleKeyEvent(e hook.Event, isDown bool) {
    if muteKeyboard {
        return
    }

    currentTime := time.Now().UnixNano() / int64(time.Millisecond)

    if isDown {
        // If the key is already pressed, don't play the sound again
        if keyState[e.Rawcode] {
            return
        }
        keyState[e.Rawcode] = true

        if currentTime-lastKeyDownTime < MULTI_KEY_DELAY {
            return
        }
        lastKeyDownTime = currentTime
        lastKeyPressTime[e.Rawcode] = currentTime

        var soundKey string
        switch e.Rawcode {
        case ENTER:
            soundKey = "down_enter"
        case SPACE:
            soundKey = "down_space"
        default:
            soundKey = getRandomDownKey()
        }
        scheduleSound(soundKey, keyboardVolume)
    } else {
        // If the key wasn't pressed (according to our state), don't play the up sound
        if !keyState[e.Rawcode] {
            return
        }
        keyState[e.Rawcode] = false

        if currentTime-lastKeyUpTime < MULTI_KEY_DELAY {
            return
        }
        lastKeyUpTime = currentTime

        lastPress, exists := lastKeyPressTime[e.Rawcode]
        if !exists || currentTime-lastPress < THRESHOLD_MS {
            return
        }

        var soundKey string
        switch e.Rawcode {
        case ENTER:
            soundKey = "up_enter"
        case SPACE:
            soundKey = "up_space"
        default:
            soundKey = getRandomUpKey()
        }
        scheduleSound(soundKey, keyboardVolume)
    }
}

func registerMouseHandlers() {
    hook.Register(hook.MouseHold, []string{}, func(e hook.Event) {
        handleMouseHold(e)
    })

    hook.Register(hook.MouseDown, []string{}, func(e hook.Event) {
        handleMouseDown(e)
    })

    hook.Register(hook.MouseUp, []string{}, func(e hook.Event) {
        handleMouseUp(e)
    })
}

func handleMouseHold(e hook.Event) {
    mouseMutex.Lock()
    defer mouseMutex.Unlock()

    if isMouseMuted {
        return
    }

    currentTime := time.Now().UnixNano() / int64(time.Millisecond)
    lastDown, exists := lastMouseDownTime[e.Button]

    if !exists || currentTime-lastDown >= THRESHOLD_MS {
        lastMouseDownTime[e.Button] = currentTime
        scheduleSound("down_mouse", mouseVolume)
    }
}

func handleMouseDown(e hook.Event) {
    mouseMutex.Lock()
    defer mouseMutex.Unlock()

    if isMouseMuted {
        return
    }

    currentTime := time.Now().UnixNano() / int64(time.Millisecond)
    lastDown, exists := lastMouseDownTime[e.Button]

    if !exists || currentTime-lastDown >= THRESHOLD_MS {
        lastMouseDownTime[e.Button] = currentTime
        scheduleSound("up_mouse", mouseVolume)
    }
}

func handleMouseUp(e hook.Event) {
    mouseMutex.Lock()
    defer mouseMutex.Unlock()

    if isMouseMuted {
        return
    }

    currentTime := time.Now().UnixNano() / int64(time.Millisecond)
    lastUp, exists := lastMouseUpTime[e.Button]

    if !exists || currentTime-lastUp >= THRESHOLD_MS {
        lastDown, downExists := lastMouseDownTime[e.Button]
        if downExists && currentTime-lastDown >= THRESHOLD_MS {
            lastMouseUpTime[e.Button] = currentTime
            scheduleSound("up_mouse", mouseVolume)
        }
    }
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
        if !keyMap[e.Button] {
            scheduleSound("down_mouse", mouseVolume)
            keyMap[e.Button] = true
        }
    })

    hook.Register(hook.MouseUp, []string{}, func(e hook.Event) {
        if keyMap[e.Button] {
            keyMap[e.Button] = false
            scheduleSound("up_mouse", mouseVolume)
        }
    })
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
    // context, err = oto.NewContext(48000, 2, 2, 8192)
    if err != nil {
        log.Fatalf("failed to create Oto context: %v", err)
    }
    // defer context.Close()

    registerKeyboardHandlers()
    registerMouseHandlers()
    isMouseMuted = muteMouse

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
    parts := strings.SplitN(command, " ", 2)

    switch parts[0] {
    case "set_keyboard", "set_mouse":
        if len(parts) < 2 {
            conn.Write([]byte(fmt.Sprintf("Invalid command. Usage: %s <new_name> [-sp <new_path>]\n", parts[0])))
            return
        }

        remainingParts := strings.SplitN(parts[1], " -sp ", 2)
        newName := strings.TrimSpace(remainingParts[0])
        newPath := ""

        if len(remainingParts) > 1 {
            newPath = strings.TrimSpace(remainingParts[1])
        }

        if parts[0] == "set_keyboard" {
            setKeyboard(conn, newName, newPath)
        } else {
            setMouse(conn, newName, newPath)
        }
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