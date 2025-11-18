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
		"welcome":
			handle_connect(data)
		"world_state":
			handle_world_state(data)
		"combat_event":
			handle_combat_event(data)
		"combat_events_batch":
			handle_combat_batch(data)
		_:
			print("Unknown message type: ", msg_type)

func handle_connect(data: Dictionary):
	if data.has("data") and data["data"].has("your_id"):
		player_id = data["data"]["your_id"]
		print("Received player ID: ", player_id)
		connection_established.emit(player_id)

func handle_world_state(data: Dictionary):
	if data.has("data"):
		var world_data = data["data"]
		world_state_received.emit(world_data)

func handle_combat_event(data: Dictionary):
	if data.has("data"):
		var event = data["data"]
		# Forward to CombatEventManager
		if CombatEventManager:
			CombatEventManager.handle_combat_event(event)

func handle_combat_batch(data: Dictionary):
	if data.has("data") and data["data"].has("events"):
		var events = data["data"]["events"]
		# Forward each event to CombatEventManager
		if CombatEventManager:
			for event in events:
				CombatEventManager.handle_combat_event(event)

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

func send_attack(target_id: String, ability_id: String = "basic_attack"):
	if not connected:
		print("[NetworkManager] Cannot attack - not connected to server")
		return

	if target_id == "":
		print("[NetworkManager] Cannot attack - no target")
		return

	var message = {
		"type": "attack",
		"player_id": player_id,
		"data": {
			"target_id": target_id,
			"ability_id": ability_id
		}
	}

	var json_string = JSON.stringify(message)
	socket.send_text(json_string)
	print("[NetworkManager] Sent attack: ", target_id, " with ", ability_id)

func send_ability(target_id: String, ability_id: String):
	if not connected:
		print("[NetworkManager] Cannot use ability - not connected to server")
		return

	var message = {
		"type": "ability",
		"player_id": player_id,
		"data": {
			"target_id": target_id,
			"ability_id": ability_id
		}
	}

	var json_string = JSON.stringify(message)
	socket.send_text(json_string)
	print("[NetworkManager] Sent ability: ", ability_id, " on ", target_id)

func send_chat(message_text: String):
	if not connected:
		return

	var message = {
		"type": "chat",
		"player_id": player_id,
		"data": {
			"message": message_text
		}
	}

	var json_string = JSON.stringify(message)
	socket.send_text(json_string)

func disconnect_from_server():
	if connected:
		socket.close()
		connected = false
