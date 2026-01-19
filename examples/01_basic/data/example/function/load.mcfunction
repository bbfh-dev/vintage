say Hello World!

execute as @a at @s run function ./_player_tick
	say 123
	tp @s ~ ~1 ~
	execute if score @s test matches 1 run function ./_has_score
		spawnpoint 0 0 0
		kill @s
	say This is the end of _player_tick!

say This is the main file
