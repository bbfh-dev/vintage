scoreboard players set #%[name] %[objective] 0
function ./_for_%[name]
	%[...]
	scoreboard players add #%[name] %[objective] 1
	execute if score #%[name] %[objective] matches %[range] run function ./_for_%[name]
