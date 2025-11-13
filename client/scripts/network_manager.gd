extends Node

# WebSocket connection
var socket = WebSocketPeer.new()
var server_url = "ws://localhost:8080/ws"

# Connection state
var connected = false
var player_id = ""

# Signals
signal connection_established(id: String)
signal connection_failed()
signal connection_closed()
signal world_state_received(players: Array)

func _ready():
	connect_to_server()

func connect_to_server():
	print("Connecting to server: ", server_url)
	var err = socket.connect_to_url(server_url)
	if err != OK:
		print("Failed to connect: ", err)
		connection_failed.emit()
		return
	print("Connection initiated...")

func _process(_delta):
	socket.poll()
	var state = socket.get_ready_state()

	# Check connection state changes
	if state == WebSocketPeer.STATE_OPEN and not connected:
		connected = true
		print("Connected to server!")
	elif state == WebSocketPeer.STATE_CLOSED and connected:
		connected = false
		print("Disconnected from server")
		connection_closed.emit()

	# Process incoming messages
	if state == WebSocketPeer.STATE_OPEN:
		while socket.get_available_packet_count() > 0:
			var packet = socket.get_packet()
			handle_server_message(packet)

func handle_server_message(packet: PackedByteArray):
	var json_string = packet.get_string_from_utf8()
	var json = JSON.new()
	var parse_result = json.parse(json_string)

	if parse_result != OK:
		print("Error parsing JSON: ", json_string)
		return

	var data = json.data
	if typeof(data) != TYPE_DICTIONARY:
		print("Invalid message format")
		return

	var msg_type = data.get("type", "")

	match msg_type:
		"connect":
			handle_connect(data)
		"world_state":
			handle_world_state(data)
		_:
			print("Unknown message type: ", msg_type)

func handle_connect(data: Dictionary):
	if data.has("data") and data["data"].has("your_id"):
		player_id = data["data"]["your_id"]
		print("Received player ID: ", player_id)
		connection_established.emit(player_id)

func handle_world_state(data: Dictionary):
	if data.has("data") and data["data"].has("players"):
		var players = data["data"]["players"]
		world_state_received.emit(players)

func send_position(position: Vector3, rotation: Vector3):
	if not connected:
		return

	var message = {
		"type": "move",
		"player_id": player_id,
		"data": {
			"position": {
				"x": position.x,
				"y": position.y,
				"z": position.z
			},
			"rotation": {
				"x": rotation.x,
				"y": rotation.y,
				"z": rotation.z
			}
		}
	}

	var json_string = JSON.stringify(message)
	socket.send_text(json_string)

func disconnect_from_server():
	if connected:
		socket.close()
		connected = false
