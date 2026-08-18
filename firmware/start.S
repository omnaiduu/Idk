.section .text.start
.global _start

_start:
    la sp, __stack_top
    call main
    ebreak

.section .bss
.align 4
__stack_bottom:
    .space 4096
__stack_top:
