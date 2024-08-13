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
    var muteKeyboard = $.getenv('keyboardVolume') === '1';
    var muteMouse = $.getenv('keyboardVolume') === '1';
    var soundsPath = $.getenv('soundsPath') || 'assets/sounds';
    var theArguments = `-ks "${keyboardSound}" -ms "${mouseSound}" -v ${globalVolume} -kv ${keyboardVolume} -mv ${mouseVolume} -sp "${soundsPath}"`;
    
    if (soundsPath !== 'assets/sounds') {
      var keyboardsPath = soundsPath + '/keyboards';
      var micePath = soundsPath + '/mice';
      
      var createdKeyboards = false;
      var createdMice = false;
  
      if (runCommand('test -d "' + keyboardsPath + '" && echo 1 || echo 0') !== '1') {
          runCommand('mkdir -p "' + keyboardsPath + '"');
          createdKeyboards = true;
      }
  
      if (runCommand('test -d "' + micePath + '" && echo 1 || echo 0') !== '1') {
          runCommand('mkdir -p "' + micePath + '"');
          createdMice = true;
      }
  
      if (createdKeyboards) {
          runCommand('cp -R "assets/sounds/keyboards/Nocfree Lite" "' + keyboardsPath + '"');
      }
  
      if (createdMice) {
          runCommand('cp -R "assets/sounds/mice/Magic Mouse" "' + micePath + '"');
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