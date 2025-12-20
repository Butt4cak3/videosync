function sanitizeRoomName(roomName) {
    return roomName
        .toLowerCase()
        .replaceAll(' ', '-')
        .replaceAll(/[^a-z0-9-_]/g, '');
}
function joinRoom(roomName) {
    roomName = sanitizeRoomName(roomName);
    window.location.href = `/room/${roomName}`;
}

document.addEventListener('DOMContentLoaded', () => {
    const nameInput = document.getElementById('room_name_input');
    const joinButton = document.getElementById('join_room_button');
    if (
        !(nameInput instanceof HTMLInputElement) ||
        !(joinButton instanceof HTMLButtonElement)
    ) {
        return;
    }

    nameInput.addEventListener('input', (event) => {
        nameInput.value = sanitizeRoomName(nameInput.value);
    });

    nameInput.addEventListener('keypress', (event) => {
        if (event.code === 'Enter') {
            joinRoom(nameInput.value);
        }
    });

    joinButton.addEventListener('click', () => {
        joinRoom(nameInput.value);
    });
});
