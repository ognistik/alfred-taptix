function run(argv) {
    const app = Application.currentApplication();
    app.includeStandardAdditions = true;

    function runCommand(command) {
        try {
            return app.doShellScript(`echo "${command}" | nc localhost 8080`).trim();
        } catch (error) {
            return null;
        }
    }

    const isAppRunning = runCommand("get_keyboard") !== null;

    if (!isAppRunning) {
        return JSON.stringify({
            items: [{
                type: 'default',
                autocomplete: 'Activate Taptix',
                title: "Activate Taptix",
                subtitle: "Taptix is Not Currently Running",
                arg: 'simple_activate',
                mods: {
					cmd: {
						valid: true,
						arg: "config",
						subtitle: "Open Taptix Configuration"
					}
                }
            }]
        });
    }

    const items = [
        {
            type: 'default',
            autocomplete: "Disable Taptix",
            title: "Disable Taptix",
            subtitle: "Turn off all sounds",
            arg: 'simple_deactivate',
            mods: {
                cmd: {
                    valid: true,
                    arg: "config",
                    subtitle: "Open Taptix Configuration"
                }
            }
        },
        {
            type: 'default',
            autocomplete: "Set Keyboard Sound",
            title: "Set Keyboard Sound",
            subtitle: runCommand("get_keyboard"),
            arg: 'set_keyboard'
        },
        {
            type: 'default',
            autocomplete: "Toggle Keyboard",
            title: "Toggle Keyboard",
            subtitle: runCommand("get_keyboard_volume") === "Keyboard is muted" ? "Unmute keyboard" : "Mute keyboard",
            arg: 'simple_toggle_keyboard'
        },
        {
            type: 'default',
            autocomplete: "Set Mouse Sound",
            title: "Set Mouse Sound",
            subtitle: runCommand("get_mouse"),
            arg: 'set_mouse'
        },
        {
            type: 'default',
            autocomplete: "Toggle Mouse",
            title: "Toggle Mouse",
            subtitle: runCommand("get_mouse_volume") === "Mouse is muted" ? "Unmute mouse" : "Mute mouse",
            arg: 'simple_toggle_mouse'
        },
        {
            type: 'default',
            autocomplete: "Set Volume",
            title: "Set Volume",
            subtitle: runCommand("get_volume"),
            arg: 'set_volume',
            mods: {
					cmd: {
						valid: true,
						arg: "set_keyboard_volume",
						subtitle: runCommand("get_keyboard_volume")
					},
                    alt: {
						valid: true,
						arg: "set_mouse_volume",
						subtitle: runCommand("get_mouse_volume")
					}
                }
        }
    ];

    return JSON.stringify({ items });
}