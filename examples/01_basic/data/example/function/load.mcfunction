execute as @a at @s run function ./_player_tick
	say 123
	tp @s ~ ~1 ~
	execute if score @s test matches 1 run function ./_has_score
		spawnpoint 0 0 0
		kill @s
	say This is the end of _player_tick!

#!calc (15 * 19) / 2
	function ./nested_a
		say deeper
	say inside of calc

#!for @s my_scoreboard ..2
	function ./nested_b
		say deeper
	say inside of calc

say This is the main file
