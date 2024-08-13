ObjC.import('stdlib');
ObjC.import('Foundation');

function run(argv) {
    var query = argv[0];
    var theAction = $.getenv('theAction');
    var soundsPath = $.getenv('soundsPath') || 'assets/sounds';

    let items = [];

    if (theAction === 'set_keyboard' || theAction === 'set_mouse') {
        const deviceType = theAction === 'set_keyboard' ? 'keyboard' : 'mouse';
        const folderPath = `${soundsPath}/${deviceType === 'mouse' ? 'mice' : 'keyboards'}`;
        const currentDevice = runCommand(`echo "get_${deviceType}" | nc localhost 8080`).replace(`Current ${deviceType}: `, '');
        
        const folders = listFolders(folderPath);
        
        if (folders.length === 0 || (folders.length === 1 && folders[0] === currentDevice)) {
            items.push({
                uid: `no_other_${deviceType}s`,
                type: 'default',
                title: `No Other ${deviceType.charAt(0).toUpperCase() + deviceType.slice(1)} Found`,
                subtitle: `Current ${deviceType}: ${currentDevice}`,
                arg: '',
            });
        } else {
            folders.forEach(folder => {
                if (folder !== currentDevice) {
                    items.push({
                        uid: `${deviceType}_${folder}`,
                        type: 'default',
                        title: folder,
                        subtitle: `USE IN TAPTIX | Currently Using ${currentDevice}`,
                        arg: folder,
                    });
                }
            });
        }
    } else if (theAction === 'set_volume' || theAction === 'set_keyboard_volume' || theAction === 'set_mouse_volume') {
        const volumeType = theAction.replace('set_', '').replace('_volume', '');
        const title = isNaN(query) || query < 0 || query > 10 
            ? `Set ${volumeType.charAt(0).toUpperCase() + volumeType.slice(1)} Volume: Choose a Value Between 0 & 10`
            : `Set ${volumeType.charAt(0).toUpperCase() + volumeType.slice(1)} Volume to ${query}`;
        
        items.push({
            uid: `set_${volumeType}_volume`,
            type: 'default',
            title: title,
            subtitle: runCommand(`echo "get_${theAction.replace('set_', '')}" | nc localhost 8080`),
            arg: query,
        });
    }

    return JSON.stringify({ items: items });
}

function listFolders(path) {
    const command = `find "${path}" -type d -depth 1 -not -path '*/.*' -exec basename {} \\;`;
    const result = runCommand(command);
    return result ? result.split('\n') : [];
}

function runCommand(cmd) {
    var task = $.NSTask.alloc.init;
    var pipe = $.NSPipe.pipe;
    task.setLaunchPath("/bin/sh");
    task.setArguments(["-c", cmd]);
    task.setStandardOutput(pipe);
    task.launch;
    var data = pipe.fileHandleForReading.readDataToEndOfFile;
    var result = $.NSString.alloc.initWithDataEncoding(data, $.NSUTF8StringEncoding).js;
    return result.trim();
}