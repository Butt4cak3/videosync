let player;
/** @type WebSocket */
let ws;
let serverState = { state: null, position: null };
let syncing = false;

document.addEventListener("DOMContentLoaded", () => {
    const input = document.getElementById("video_url_input");
    const button = document.getElementById("play_video_button");
    if (
        !(input instanceof HTMLInputElement) ||
        !(button instanceof HTMLButtonElement)
    ) {
        return;
    }

    input.addEventListener("keypress", (event) => {
        if (event.code === "Enter") {
            playVideo(input.value);
        }
    });

    button.addEventListener("click", () => {
        playVideo(input.value);
    });
});

function playVideo(url) {
    ws.send(
        JSON.stringify({
            type: "loadurl",
            payload: {
                url,
            },
        })
    );
}

function createPlayer(events) {
    return new Promise((resolve) => {
        const player = new YT.Player("yt_player", {
            width: 640,
            height: 360,
            events: {},
            events: {
                ...events,
                onReady: () => {
                    resolve(player);
                },
            },
        });
    });
}

function connectSocket(roomId) {
    const ws = new WebSocket(`ws://${document.location.host}/socket/${roomId}`);
    return ws;
}

function updatePlayerState(newState) {
    serverState = { ...newState };

    syncing = true;
    setTimeout(() => {
        syncing = false;
    }, 250);

    switch (newState.state) {
        case YT.PlayerState.PLAYING:
            player.playVideo();
            break;
        case YT.PlayerState.PAUSED:
            player.pauseVideo();
            break;
    }

    player.seekTo(newState.position);
}

async function initRoom(roomId) {
    player = await createPlayer({
        onStateChange: (event) => {
            if (syncing) {
                return;
            }

            switch (event.data) {
                case YT.PlayerState.PLAYING:
                    if (serverState.state !== YT.PlayerState.PLAYING) {
                        serverState.state = YT.PlayerState.PLAYING;
                        serverState.position = player.getCurrentTime();
                        ws.send(
                            JSON.stringify({
                                type: "play",
                                payload: {
                                    position: player.getCurrentTime(),
                                },
                            })
                        );
                    }
                    break;
                case YT.PlayerState.PAUSED:
                    if (serverState.state !== YT.PlayerState.PAUSED) {
                        console.log(
                            `Sending pause because serverState.state is ${serverState.state}`
                        );
                        serverState.state = YT.PlayerState.PAUSED;
                        serverState.position = player.getCurrentTime();
                        ws.send(
                            JSON.stringify({
                                type: "pause",
                                payload: {
                                    position: player.getCurrentTime(),
                                },
                            })
                        );
                    }
                    break;
            }
        },
    });

    player.addEventListener("onPlaying", () => {
        console.log("playing");
    });

    ws = connectSocket(roomId);

    ws.addEventListener("message", (event) => {
        const { type, payload } = JSON.parse(event.data);
        switch (type) {
            case "init":
                player.loadVideoById(payload.videoId, payload.videoPos);
                if (payload.playbackState == YT.PlayerState.PAUSED) {
                    player.pauseVideo();
                }
                serverState = {
                    state: payload.playbackState,
                    position: payload.videoPos,
                };
                break;
            case "play":
                updatePlayerState({
                    state: YT.PlayerState.PLAYING,
                    position: payload.position,
                });
                break;
            case "pause":
                updatePlayerState({
                    state: YT.PlayerState.PAUSED,
                    position: payload.position,
                });
                break;
            case "load":
                player.loadVideoById(payload.videoId, 0);
                updatePlayerState({
                    state: YT.PlayerState.PAUSED,
                    position: 0,
                });
        }
    });

    setInterval(() => {
        if (syncing) {
            return;
        }

        const state = player.getPlayerState();
        if (state === YT.PlayerState.PAUSED) {
            const currentTime = player.getCurrentTime();
            if (
                serverState.position !== null &&
                currentTime !== serverState.position
            ) {
                serverState.position = currentTime;
                ws.send(
                    JSON.stringify({
                        type: "pause",
                        payload: {
                            position: currentTime,
                        },
                    })
                );
            }
            prevTime = currentTime;
        } else {
            prevTime = -1;
        }
    }, 500);
}
