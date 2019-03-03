Z80EMU_HOME=$(HOME)/z80emu

all: z80emu/icmp-z80emu
	z80emu/icmp-z80emu > z80emu.txt
	go run icmp-rcs/icmp-rcs.go > rcs.txt
	diff z80emu.txt rcs.txt

z80emu/icmp-z80emu: z80emu/icmp-z80emu.c
	gcc -o z80emu/icmp-z80emu \
		-I$(Z80EMU_HOME) \
		$(Z80EMU_HOME)/z80emu.c \
		z80emu/icmp-z80emu.c

clean:
	rm *.txt
	rm z80emu/icmp-z80emu

