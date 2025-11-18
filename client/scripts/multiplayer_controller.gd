extends Node3D

# Scenes to instantiate
@export var player_scene: PackedScene
@export var monster_scene: PackedScene

# Track spawned entities
var spawned_players = {}
var spawned_monsters = {}
var local_player = null
var local_player_id = ""

func _ready():
	# Connect to network signals
	NetworkManager.connection_established.connect(_on_connection_established)
	NetworkManager.world_state_received.connect(_on_world_state_received)

func _on_connection_established(id: String):
	print("Connection established with ID: ", id)
	local_player_id = id
	spawn_local_player(id)

func spawn_local_player(id: String):
	if player_scene == null:
		print("Error: Player scene not assigned!")
		return

	var player = player_scene.instantiate()
	player.is_local_player = true
	player.player_id = id
	player.position = Vector3(0, 1, 0)
	player.name = "LocalPlayer_" + id
	add_child(player)
	local_player = player

	print("Local player spawned: ", id)

func spawn_remote_player(id: String, pos: Vector3, rot: Vector3):
	if player_scene == null:
		return

	var player = player_scene.instantiate()
	player.is_local_player = false
	player.player_id = id
	player.position = pos
	player.rotation = rot
	player.name = "RemotePlayer_" + id
	add_child(player)
	spawned_players[id] = player

	print("Remote player spawned: ", id)

func _on_world_state_received(world_data: Dictionary):
	# Handle players
	var players = world_data.get("players", [])
	for player_data in players:
		var id = player_data.get("id", "")
		var pos_data = player_data.get("position", {})
		var rot_data = player_data.get("rotation", {})

		var pos = Vector3(
			pos_data.get("x", 0),
			pos_data.get("y", 0),
			pos_data.get("z", 0)
		)
		var rot = Vector3(
			rot_data.get("x", 0),
			rot_data.get("y", 0),
			rot_data.get("z", 0)
		)

		# Skip local player
		if id == local_player_id:
			continue

		# Update or spawn remote player
		if spawned_players.has(id):
			spawned_players[id].set_remote_transform(pos, rot)
		else:
			spawn_remote_player(id, pos, rot)

	# Remove disconnected players
	var current_player_ids = []
	for player_data in players:
		current_player_ids.append(player_data.get("id", ""))

	for id in spawned_players.keys():
		if id not in current_player_ids:
			print("Removing disconnected player: ", id)
			spawned_players[id].queue_free()
			spawned_players.erase(id)

	# Handle monsters
	var monsters = world_data.get("monsters", [])
	for monster_data in monsters:
		var id = monster_data.get("id", "")
		var name = monster_data.get("name", "Monster")
		var pos_data = monster_data.get("position", {})
		var health = monster_data.get("health", 100)
		var max_health = monster_data.get("max_health", 100)

		var pos = Vector3(
			pos_data.get("x", 0),
			pos_data.get("y", 0),
			pos_data.get("z", 0)
		)

		# Update or spawn monster
		if spawned_monsters.has(id):
			spawned_monsters[id].update_from_server(pos, health, max_health)
		else:
			spawn_monster(id, name, pos, health, max_health)

	# Remove dead/despawned monsters
	var current_monster_ids = []
	for monster_data in monsters:
		current_monster_ids.append(monster_data.get("id", ""))

	for id in spawned_monsters.keys():
		if id not in current_monster_ids:
			print("Removing despawned monster: ", id)
			spawned_monsters[id].queue_free()
			spawned_monsters.erase(id)

func spawn_monster(id: String, name: String, pos: Vector3, health: int, max_hp: int):
	if monster_scene == null:
		print("Warning: Monster scene not assigned!")
		return

	var monster = monster_scene.instantiate()
	monster.entity_id = id
	monster.monster_name = name
	monster.current_health = health
	monster.max_health = max_hp
	monster.position = pos
	monster.name = "Monster_" + id
	add_child(monster)
	spawned_monsters[id] = monster

	print("Monster spawned: ", id, " (", name, ")")
