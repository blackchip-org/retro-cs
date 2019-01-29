#include <stdio.h>

#include "z80emu.h"
#include "z80user.h"

typedef struct ix {
    Z80_STATE	    state;
	unsigned char	memory[1 << 16];
} ix;

void cpd() {
    struct ix context;
    for (int a = 0; a <= 0xff; a++) {
        for (int _hl = 0; _hl <= 0xff; _hl++) {
            for (int f = 0; f <= 1; f++) {
                for (int c = 2; c <= 2; c++) {
                    Z80Reset(&context.state);
                    context.memory[0x1234] = _hl;
                    context.state.registers.byte[Z80_B] = 0;
                    context.state.registers.byte[Z80_C] = c;
                    context.state.registers.byte[Z80_H] = 0x12;
                    context.state.registers.byte[Z80_L] = 0x34;
                    context.state.registers.byte[Z80_A] = a;
                    context.state.registers.byte[Z80_F] = f;
                    context.memory[0] = 0xed;
                    context.memory[1] = 0xb9;
                    Z80Emulate(&context.state, 0, &context);
                    printf("a:%02x _hl:%02x c:%02x f:%02x => b:%02x c:%02x f:%02x\n", a, _hl, c, f,
                        context.state.registers.byte[Z80_B],
                        context.state.registers.byte[Z80_C],
                        context.state.registers.byte[Z80_F]
                    );
                }
            }
        }
    }
}

void af() {
    struct ix context;
    for (int f = 0; f <= 0xff; f++) {
        for (int a = 0; a <= 0xff; a++) {
            Z80Reset(&context.state);
            context.memory[0] = 0x17;
            context.state.registers.byte[Z80_A] = a;
            context.state.registers.byte[Z80_F] = f;
            Z80Emulate(&context.state, 0, &context);
            printf("a:%02x f:%02x => a:%02x f:%02x\n", a, f,
                context.state.registers.byte[Z80_A],
                context.state.registers.byte[Z80_F]
            );
        }
    }
}

void SystemCall (ZEXTEST *zextest) {}

int main() {
    af();
}
