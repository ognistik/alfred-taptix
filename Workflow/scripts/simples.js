ObjC.import('stdlib');
ObjC.import('Foundation');

function run(argv) {
  var theAction = $.getenv('theAction');

  if (theAction === 'simple_toggle_mouse') {
    return runCommand('echo "toggle_mouse" | nc localhost 8080');
  } else if (theAction === 'simple_toggle_keyboard') {
    return runCommand('echo "toggle_keyboard" | nc localhost 8080');
  } else if (theAction === 'simple_deactivate') {
    runCommand('sleep 0.3 && echo "quit" | nc localhost 8080');
    return 'Taptix has been deactivated';
  } else if (theAction === 'simple_activate') {
    var keyboardSound = $.getenv('keyboardSound');
    var mouseSound = $.getenv('mouseSound');
    var globalVolume = $.getenv('globalVolume');
    var keyboardVolume = $.getenv('keyboardVolume');
    var mouseVolume = $.getenv('mouseVolume');
    var muteKeyboard = $.getenv('muteKeyboard') === '1';
    var muteMouse = $.getenv('muteMouse') === '1';
    var bothPaths = $.getenv('bothPaths') === '1';
    var soundsPath = $.getenv('soundsPath');

    var keyboardsPath, micePath;

    if (bothPaths) {
        if (!soundsPath) {
            keyboardsPath = 'assets/sounds';
            micePath = 'assets/sounds';
        } else {
            var defaultKeyboardPath = 'assets/sounds/keyboards/' + keyboardSound;
            var defaultMousePath = 'assets/sounds/mice/' + mouseSound;
            
            keyboardsPath = runCommand('test -d "' + defaultKeyboardPath + '" && echo 1 || echo 0') === '1' ? 'assets/sounds' : soundsPath;
            micePath = runCommand('test -d "' + defaultMousePath + '" && echo 1 || echo 0') === '1' ? 'assets/sounds' : soundsPath;
        }
    } else {
        keyboardsPath = soundsPath || 'assets/sounds';
        micePath = soundsPath || 'assets/sounds';
    }

    var theArguments = `-ks "${keyboardSound}" -ms "${mouseSound}" -v ${globalVolume} -kv ${keyboardVolume} -mv ${mouseVolume} -kp "${keyboardsPath}" -mp "${micePath}"`;

    if (keyboardsPath !== 'assets/sounds') {
        var fullKeyboardsPath = keyboardsPath + '/keyboards';
        if (runCommand('test -d "' + fullKeyboardsPath + '" && echo 1 || echo 0') !== '1') {
            runCommand('mkdir -p "' + fullKeyboardsPath + '"');
            runCommand('cp -R "assets/sounds/keyboards/Nocfree Lite" "' + fullKeyboardsPath + '"');
        }
    }

    if (micePath !== 'assets/sounds') {
        var fullMicePath = micePath + '/mice';
        if (runCommand('test -d "' + fullMicePath + '" && echo 1 || echo 0') !== '1') {
            runCommand('mkdir -p "' + fullMicePath + '"');
            runCommand('cp -R "assets/sounds/mice/Magic Mouse" "' + fullMicePath + '"');
        }
    }

    if (muteKeyboard) {
        theArguments += ' -mk';
    }
    if (muteMouse) {
        theArguments += ' -mm';
    }

    return JSON.stringify({
        alfredworkflow: {
            arg: "activate",
            variables: {
                theKeyboard: keyboardSound,
                theMouse: mouseSound,
                theArguments: theArguments
            }
        }
    });
  }

  return 'Unknown command';
}

function runCommand(command) {
  var task = $.NSTask.alloc.init;
  task.setLaunchPath("/bin/bash");
  task.setArguments(["-c", command]);

  var pipe = $.NSPipe.pipe;
  task.standardOutput = pipe;
  task.standardError = pipe;

  var fileHandle = pipe.fileHandleForReading;
  task.launch;

  var data = fileHandle.readDataToEndOfFile;
  var result = $.NSString.alloc.initWithDataEncoding(data, $.NSUTF8StringEncoding).js;

  return result.trim();
}