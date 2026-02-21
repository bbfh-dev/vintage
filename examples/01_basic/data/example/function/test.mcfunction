function ./_greet
	say "Hello World!"
	data merge entity @s { \
		Silent: 1b, \
		Tags: [ \
			"testing_indentation" \
		] \
	}

execute as @e \
	at @s \
	run function bs.math:some_function/here
