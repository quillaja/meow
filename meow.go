// +build amd64,cgo

// Package meow uses cgo to wrap the "meow hash" developed by
// Molly Rocket, Inc. It implements a non-cryptographic hash.
// See header comment in cgo section of meow.go for more details,
// or the repo at https://github.com/cmuratori/meow_hash.
//
// One will primarily want to use the Hash() function. This package
// also provides compatibility with the hash.Hash interface using New().
package meow

// #cgo CFLAGS: -O3 -mavx -maes
//
// /* ========================================================================
//
//    Meow - A Fast Non-cryptographic Hash
//    (C) Copyright 2018-2019 by Molly Rocket, Inc. (https://mollyrocket.com)
//
//    See https://mollyrocket.com/meowhash for details.
//
//    ========================================================================
//
//    zlib License
//
//    (C) Copyright 2018-2019 Molly Rocket, Inc.
//
//    This software is provided 'as-is', without any express or implied
//    warranty.  In no event will the authors be held liable for any damages
//    arising from the use of this software.
//
//    Permission is granted to anyone to use this software for any purpose,
//    including commercial applications, and to alter it and redistribute it
//    freely, subject to the following restrictions:
//
//    1. The origin of this software must not be misrepresented; you must not
//       claim that you wrote the original software. If you use this software
//       in a product, an acknowledgment in the product documentation would be
//       appreciated but is not required.
//    2. Altered source versions must be plainly marked as such, and must not be
//       misrepresented as being the original software.
//    3. This notice may not be removed or altered from any source distribution.
//
//    ========================================================================
//
//    FAQ
//
//    Q: What is it?
//
//    A: Meow is a 128-bit Level 3 hash taking 128 bytes of seed.  It operates
//       at very high speeds on x64 processors, and potentially other processors
//       that provide accelerated AES instructions.
//
//    Q: What is it GOOD for?
//
//    A: Quickly hashing any amount of data for comparison purposes such as
//       block deduplication or change detection.  It is fast on all buffer
//       sizes, and can generally be used anywhere you need fast Level 3
//       hashing without worrying about how big or small the inputs tend to be.
//
//       However, substantial speed improvements could be made over Meow
//       if you either a) know you are always hashing an exact, small number of bytes,
//       or b) can always supply a small number of bytes in a buffer padded to some
//       fixed multiple of 16.
//
//    Q: What is it BAD for?
//
//    A: Anything requiring Level 4 or Level 5 security guarantees (see
//       http://nohatcoder.dk/2019-05-19-1.html#level3).  Also, note that
//       Meow is a new hash and has not had the extensive community
//       cryptanalysis necessary to ensure that it is not breakable down to
//       a lower level of hash, so you must do your due diligence in
//       deciding when and where to use Meow instead of a slower but
//       more extensively studied existing hash.  We have tried to design
//       it to provide Level 3 security, but the possibility of the hash
//       being broken in the future always exists.
//
//    Q: Why is it called the "Meow hash"?
//
//    A: It is named after a character in Meow the Infinite
//       (https://meowtheinfinite.com)
//
//    Q: Who wrote it?
//
//    A: The final Meow Hash was created as a collaboration between
//       JACOB CHRISTIAN MUNCH-ANDERSEN (https://twitter.com/nohatcoder) and
//       CASEY MURATORI (https://caseymuratori.com).  Casey wrote the original
//       implementation for use in processing large-footprint assets for the
//       game 1935 (https://molly1935.com).  Jacob was the first to analyze
//       that implementation and determine the adversarial bit strength, which
//       was weaker than they would have liked.
//
//       Following that, the two collaborated to figure out how the hash
//       could be strengthened without reducing Meow's 16 bytes/cycle
//       maximum theoretical throughput.  Jacob created the hash candidates
//       and Casey did the performance validation.  After a long and
//       exhaustive effort, Jacob found the unaligned aes/add/xor formulation
//       that forms the current Meow hash core.
//
//       A number of valuable additions to Meow Hash were also contributed
//       by other great folks along the way:
//
//       JEFF ROBERTS (https://radgametools.com) provided a super slick
//       way to handle the residual end-of-buffer bytes that dramatically
//       improved Meow's small hash performance.
//
//       MARTINS MOZEIKO (https://matrins.ninja) ported Meow to ARM and
//       ANSI-C, and added the proper preprocessor dressing for clean
//       compilation on a variety of compiler configurations.
//
//       FABIAN GIESEN (https://fgiesen.wordpress.com) analyzed many
//       performance oddities that came up during development, and
//       helped get the benchmarking working properly across a number
//       of platforms.
//
//       ARAS PRANCKEVICIUS (https://aras-p.info) provided the allocation
//       shim for compilation on Mac OS X.
//
//    ======================================================================== */
//
// //
// // IMPORTANT(casey): We are currently evaluating this hash construction as
// // the final one for Meow Hash.  If you find a way to produce collisions
// // that should not be possible with a Level 3 hash, find significant performance
// // problems, or see any bugs in this version, please be sure to report them
// // to the Meow Hash GitHub as soon as possible.  We would like to know as
// // much as we can about the robustness and performance before committing to
// // it as the final construction.
// //
//
// #if !defined(MEOW_HASH_X64_AESNI_H)
//
// #define MEOW_HASH_VERSION 5
// #define MEOW_HASH_VERSION_NAME "0.5/calico"
//
// #if !defined(meow_u8)
//
// #if _MSC_VER
// #if !defined(__clang__)
// #define INSTRUCTION_REORDER_BARRIER _ReadWriteBarrier()
// #else
// #endif
// #include <intrin.h>
// #else
// #include <x86intrin.h>
// #endif
//
// #define meow_u8 char unsigned
// #define meow_u64 long long unsigned
// #define meow_u128 __m128i
//
// #if __x86_64__ || _M_AMD64
// #define meow_umm long long unsigned
// #define MeowU64From(A, I) (_mm_extract_epi64((A), (I)))
// #elif __i386__  || _M_IX86
// #define meow_umm int unsigned
// #define MeowU64From(A, I) (*(meow_u64 *)&(A))
// #else
// #error Cannot determine architecture to use!
// #endif
//
// #define MeowU32From(A, I) (_mm_extract_epi32((A), (I)))
// #define MeowHashesAreEqual(A, B) (_mm_movemask_epi8(_mm_cmpeq_epi8((A), (B))) == 0xFFFF)
//
// #if !defined INSTRUCTION_REORDER_BARRIER
// #define INSTRUCTION_REORDER_BARRIER
// #endif
//
// #if !defined MEOW_PAGESIZE
// #define MEOW_PAGESIZE 4096
// #endif
//
// #if !defined MEOW_PREFETCH
// #define MEOW_PREFETCH 4096
// #endif
//
// #if !defined MEOW_PREFETCH_LIMIT
// #define MEOW_PREFETCH_LIMIT 0x3ff
// #endif
//
// #endif
//
// #define prefetcht0(A) _mm_prefetch((char *)(A), _MM_HINT_T0)
// #define movdqu(A, B)  A = _mm_loadu_si128((__m128i *)(B))
// #define movdqu_mem(A, B)  _mm_storeu_si128((__m128i *)(A), B)
// #define movq(A, B) A = _mm_set_epi64x(0, B);
// #define aesdec(A, B)  A = _mm_aesdec_si128(A, B)
// #define pshufb(A, B)  A = _mm_shuffle_epi8(A, B)
// #define pxor(A, B)    A = _mm_xor_si128(A, B)
// #define paddq(A, B) A = _mm_add_epi64(A, B)
// #define pand(A, B)    A = _mm_and_si128(A, B)
// #define palignr(A, B, i) A = _mm_alignr_epi8(A, B, i)
// #define pxor_clear(A, B)    A = _mm_setzero_si128(); // NOTE(casey): pxor_clear is a nonsense thing that is only here because compilers don't detect xor(a, a) is clearing a :(
//
// #define MEOW_MIX_REG(r1, r2, r3, r4, r5,  i1, i2, i3, i4) \
// aesdec(r1, r2); \
// INSTRUCTION_REORDER_BARRIER; \
// paddq(r3, i1); \
// pxor(r2, i2); \
// aesdec(r2, r4); \
// INSTRUCTION_REORDER_BARRIER; \
// paddq(r5, i3); \
// pxor(r4, i4);
//
// #define MEOW_MIX(r1, r2, r3, r4, r5,  ptr) \
// MEOW_MIX_REG(r1, r2, r3, r4, r5, _mm_loadu_si128( (__m128i *) ((ptr) + 15) ), _mm_loadu_si128( (__m128i *) ((ptr) + 0)  ), _mm_loadu_si128( (__m128i *) ((ptr) + 1)  ), _mm_loadu_si128( (__m128i *) ((ptr) + 16) ))
//
// #define MEOW_SHUFFLE(r1, r2, r3, r4, r5, r6) \
// aesdec(r1, r4); \
// paddq(r2, r5); \
// pxor(r4, r6); \
// aesdec(r4, r2); \
// paddq(r5, r6); \
// pxor(r2, r3)
//
// #if MEOW_DUMP
// struct meow_dump
// {
//     meow_u128 xmm[8];
//     void *Ptr;
//     char const *Title;
// };
// extern "C" meow_dump *MeowDumpTo;
// meow_dump *MeowDumpTo;
// #define MEOW_DUMP_STATE(T, xmm0, xmm1, xmm2, xmm3, xmm4, xmm5, xmm6, xmm7, ptr) \
// if(MeowDumpTo) \
// { \
//     MeowDumpTo->xmm[0] = xmm0; \
//     MeowDumpTo->xmm[1] = xmm1; \
//     MeowDumpTo->xmm[2] = xmm2; \
//     MeowDumpTo->xmm[3] = xmm3; \
//     MeowDumpTo->xmm[4] = xmm4; \
//     MeowDumpTo->xmm[5] = xmm5; \
//     MeowDumpTo->xmm[6] = xmm6; \
//     MeowDumpTo->xmm[7] = xmm7; \
//     MeowDumpTo->Ptr = ptr; \
//     MeowDumpTo->Title = T; \
//     ++MeowDumpTo; \
// }
// #else
// #define MEOW_DUMP_STATE(...)
// #endif
//
// static meow_u8 MeowShiftAdjust[32] = {0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15};
// static meow_u8 MeowMaskLen[32] = {255,255,255,255, 255,255,255,255, 255,255,255,255, 255,255,255,255, 0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0};
//
// // NOTE(casey): The default seed is now a "nothing-up-our-sleeves" number for good measure.  You may verify that it is just an encoding of Pi.
// static meow_u8 MeowDefaultSeed[128] =
// {
//     0x32, 0x43, 0xF6, 0xA8, 0x88, 0x5A, 0x30, 0x8D,
// 	0x31, 0x31, 0x98, 0xA2, 0xE0, 0x37, 0x07, 0x34,
// 	0x4A, 0x40, 0x93, 0x82, 0x22, 0x99, 0xF3, 0x1D,
// 	0x00, 0x82, 0xEF, 0xA9, 0x8E, 0xC4, 0xE6, 0xC8,
// 	0x94, 0x52, 0x82, 0x1E, 0x63, 0x8D, 0x01, 0x37,
// 	0x7B, 0xE5, 0x46, 0x6C, 0xF3, 0x4E, 0x90, 0xC6,
// 	0xCC, 0x0A, 0xC2, 0x9B, 0x7C, 0x97, 0xC5, 0x0D,
// 	0xD3, 0xF8, 0x4D, 0x5B, 0x5B, 0x54, 0x70, 0x91,
// 	0x79, 0x21, 0x6D, 0x5D, 0x98, 0x97, 0x9F, 0xB1,
// 	0xBD, 0x13, 0x10, 0xBA, 0x69, 0x8D, 0xFB, 0x5A,
// 	0xC2, 0xFF, 0xD7, 0x2D, 0xBD, 0x01, 0xAD, 0xFB,
// 	0x7B, 0x8E, 0x1A, 0xFE, 0xD6, 0xA2, 0x67, 0xE9,
// 	0x6B, 0xA7, 0xC9, 0x04, 0x5F, 0x12, 0xC7, 0xF9,
// 	0x92, 0x4A, 0x19, 0x94, 0x7B, 0x39, 0x16, 0xCF,
// 	0x70, 0x80, 0x1F, 0x2E, 0x28, 0x58, 0xEF, 0xC1,
// 	0x66, 0x36, 0x92, 0x0D, 0x87, 0x15, 0x74, 0xE6
// };
//
// //
// // NOTE(casey): Single block version
// //
//
// static meow_u128
// MeowHash(void *Seed128Init, meow_umm Len, void *SourceInit)
// {
//     meow_u128 xmm0, xmm1, xmm2, xmm3, xmm4, xmm5, xmm6, xmm7; // NOTE(casey): xmm0-xmm7 are the hash accumulation lanes
//     meow_u128 xmm8, xmm9, xmm10, xmm11, xmm12, xmm13, xmm14, xmm15; // NOTE(casey): xmm8-xmm15 hold values to be appended (residual, length)
//
//     meow_u8 *rax = (meow_u8 *)SourceInit;
//     meow_u8 *rcx = (meow_u8 *)Seed128Init;
//
//     //
// 	   // NOTE(casey): Seed the eight hash registers
//     //
//
//     movdqu(xmm0, rcx + 0x00);
//     movdqu(xmm1, rcx + 0x10);
//     movdqu(xmm2, rcx + 0x20);
//     movdqu(xmm3, rcx + 0x30);
//
//     movdqu(xmm4, rcx + 0x40);
//     movdqu(xmm5, rcx + 0x50);
//     movdqu(xmm6, rcx + 0x60);
//     movdqu(xmm7, rcx + 0x70);
//
//     MEOW_DUMP_STATE("Seed", xmm0, xmm1, xmm2, xmm3, xmm4, xmm5, xmm6, xmm7, 0);
//
//     //
//     // NOTE(casey): Hash all full 256-byte blocks
//     //
//
//     meow_umm BlockCount = (Len >> 8);
//     if(BlockCount > MEOW_PREFETCH_LIMIT)
//     {
//         // NOTE(casey): For large input, modern Intel x64's can't hit full speed without prefetching, so we use this loop
//         while(BlockCount--)
//         {
//             prefetcht0(rax + MEOW_PREFETCH + 0x00);
//             prefetcht0(rax + MEOW_PREFETCH + 0x40);
//             prefetcht0(rax + MEOW_PREFETCH + 0x80);
//             prefetcht0(rax + MEOW_PREFETCH + 0xc0);
//
//             MEOW_MIX(xmm0,xmm4,xmm6,xmm1,xmm2, rax + 0x00);
//             MEOW_MIX(xmm1,xmm5,xmm7,xmm2,xmm3, rax + 0x20);
//             MEOW_MIX(xmm2,xmm6,xmm0,xmm3,xmm4, rax + 0x40);
//             MEOW_MIX(xmm3,xmm7,xmm1,xmm4,xmm5, rax + 0x60);
//             MEOW_MIX(xmm4,xmm0,xmm2,xmm5,xmm6, rax + 0x80);
//             MEOW_MIX(xmm5,xmm1,xmm3,xmm6,xmm7, rax + 0xa0);
//             MEOW_MIX(xmm6,xmm2,xmm4,xmm7,xmm0, rax + 0xc0);
//             MEOW_MIX(xmm7,xmm3,xmm5,xmm0,xmm1, rax + 0xe0);
//
//             rax += 0x100;
//         }
//     }
//     else
//     {
//         // NOTE(casey): For small input, modern Intel x64's can't hit full speed _with_ prefetching (because of port pressure), so we use this loop.
//         while(BlockCount--)
//         {
//             MEOW_MIX(xmm0,xmm4,xmm6,xmm1,xmm2, rax + 0x00);
//             MEOW_MIX(xmm1,xmm5,xmm7,xmm2,xmm3, rax + 0x20);
//             MEOW_MIX(xmm2,xmm6,xmm0,xmm3,xmm4, rax + 0x40);
//             MEOW_MIX(xmm3,xmm7,xmm1,xmm4,xmm5, rax + 0x60);
//             MEOW_MIX(xmm4,xmm0,xmm2,xmm5,xmm6, rax + 0x80);
//             MEOW_MIX(xmm5,xmm1,xmm3,xmm6,xmm7, rax + 0xa0);
//             MEOW_MIX(xmm6,xmm2,xmm4,xmm7,xmm0, rax + 0xc0);
//             MEOW_MIX(xmm7,xmm3,xmm5,xmm0,xmm1, rax + 0xe0);
//
//             rax += 0x100;
//         }
//     }
//
//     MEOW_DUMP_STATE("PostBlocks", xmm0, xmm1, xmm2, xmm3, xmm4, xmm5, xmm6, xmm7, 0);
//
//     //
//     // NOTE(casey): Load any less-than-32-byte residual
//     //
//
//     pxor_clear(xmm9, xmm9);
//     pxor_clear(xmm11, xmm11);
//
//     //
//     // TODO(casey): I need to put more thought into how the end-of-buffer stuff is actually working out here,
//     // because I _think_ it may be possible to remove the first branch (on Len8) and let the mask zero out the
//     // result, but it would take a little thought to make sure it couldn't read off the end of the buffer due
//     // to the & 0xf on the align computation.
//     //
//
//     // NOTE(casey): First, we have to load the part that is _not_ 16-byte aligned
//     meow_u8 *Last = (meow_u8 *)SourceInit + (Len & ~0xf);
//     int unsigned Len8 = (Len & 0xf);
//     if(Len8)
//     {
//         // NOTE(casey): Load the mask early
//         movdqu(xmm8, &MeowMaskLen[0x10 - Len8]);
//
//         meow_u8 *LastOk = (meow_u8*)((((meow_umm)(((meow_u8 *)SourceInit)+Len - 1)) | (MEOW_PAGESIZE - 1)) - 16);
//         int Align = (Last > LastOk) ? ((int)(meow_umm)Last) & 0xf : 0;
//         movdqu(xmm10, &MeowShiftAdjust[Align]);
//         movdqu(xmm9, Last - Align);
//         pshufb(xmm9, xmm10);
//
//         // NOTE(jeffr): and off the extra bytes
//         pand(xmm9, xmm8);
//     }
//
//     // NOTE(casey): Next, we have to load the part that _is_ 16-byte aligned
//     if(Len & 0x10)
//     {
//         xmm11 = xmm9;
//         movdqu(xmm9, Last - 0x10);
//     }
//
//     //
//     // NOTE(casey): Construct the residual and length injests
//     //
//
//     xmm8 = xmm9;
//     xmm10 = xmm9;
//     palignr(xmm8, xmm11, 15);
//     palignr(xmm10, xmm11, 1);
//
//     // NOTE(casey): We have room for a 128-bit nonce and a 64-bit none here, but
//     // the decision was made to leave them zero'd so as not to confuse people
//     // about hwo to use them or what security implications they had.
//     pxor_clear(xmm12, xmm12);
//     pxor_clear(xmm13, xmm13);
//     pxor_clear(xmm14, xmm14);
//     movq(xmm15, Len);
//     palignr(xmm12, xmm15, 15);
//     palignr(xmm14, xmm15, 1);
//
//     MEOW_DUMP_STATE("Residuals", xmm8, xmm9, xmm10, xmm11, xmm12, xmm13, xmm14, xmm15, 0);
//
//     // NOTE(casey): To maintain the mix-down pattern, we always Meow Mix the less-than-32-byte residual, even if it was empty
//     MEOW_MIX_REG(xmm0, xmm4, xmm6, xmm1, xmm2,  xmm8, xmm9, xmm10, xmm11);
//
//     // NOTE(casey): Append the length, to avoid problems with our 32-byte padding
//     MEOW_MIX_REG(xmm1, xmm5, xmm7, xmm2, xmm3,  xmm12, xmm13, xmm14, xmm15);
//
//     MEOW_DUMP_STATE("PostAppend", xmm0, xmm1, xmm2, xmm3, xmm4, xmm5, xmm6, xmm7, 0);
//
//     //
//     // NOTE(casey): Hash all full 32-byte blocks
//     //
//     int unsigned LaneCount = (Len >> 5) & 0x7;
//     if(LaneCount == 0) goto MixDown; MEOW_MIX(xmm2,xmm6,xmm0,xmm3,xmm4, rax + 0x00); --LaneCount;
//     if(LaneCount == 0) goto MixDown; MEOW_MIX(xmm3,xmm7,xmm1,xmm4,xmm5, rax + 0x20); --LaneCount;
//     if(LaneCount == 0) goto MixDown; MEOW_MIX(xmm4,xmm0,xmm2,xmm5,xmm6, rax + 0x40); --LaneCount;
//     if(LaneCount == 0) goto MixDown; MEOW_MIX(xmm5,xmm1,xmm3,xmm6,xmm7, rax + 0x60); --LaneCount;
//     if(LaneCount == 0) goto MixDown; MEOW_MIX(xmm6,xmm2,xmm4,xmm7,xmm0, rax + 0x80); --LaneCount;
//     if(LaneCount == 0) goto MixDown; MEOW_MIX(xmm7,xmm3,xmm5,xmm0,xmm1, rax + 0xa0); --LaneCount;
//     if(LaneCount == 0) goto MixDown; MEOW_MIX(xmm0,xmm4,xmm6,xmm1,xmm2, rax + 0xc0); --LaneCount;
//
//     //
//     // NOTE(casey): Mix the eight lanes down to one 128-bit hash
//     //
//
//     MixDown:
//
//     MEOW_DUMP_STATE("PostLanes", xmm0, xmm1, xmm2, xmm3, xmm4, xmm5, xmm6, xmm7, 0);
//
//     MEOW_SHUFFLE(xmm0, xmm1, xmm2, xmm4, xmm5, xmm6);
//     MEOW_SHUFFLE(xmm1, xmm2, xmm3, xmm5, xmm6, xmm7);
//     MEOW_SHUFFLE(xmm2, xmm3, xmm4, xmm6, xmm7, xmm0);
//     MEOW_SHUFFLE(xmm3, xmm4, xmm5, xmm7, xmm0, xmm1);
//     MEOW_SHUFFLE(xmm4, xmm5, xmm6, xmm0, xmm1, xmm2);
//     MEOW_SHUFFLE(xmm5, xmm6, xmm7, xmm1, xmm2, xmm3);
//     MEOW_SHUFFLE(xmm6, xmm7, xmm0, xmm2, xmm3, xmm4);
//     MEOW_SHUFFLE(xmm7, xmm0, xmm1, xmm3, xmm4, xmm5);
//     MEOW_SHUFFLE(xmm0, xmm1, xmm2, xmm4, xmm5, xmm6);
//     MEOW_SHUFFLE(xmm1, xmm2, xmm3, xmm5, xmm6, xmm7);
//     MEOW_SHUFFLE(xmm2, xmm3, xmm4, xmm6, xmm7, xmm0);
//     MEOW_SHUFFLE(xmm3, xmm4, xmm5, xmm7, xmm0, xmm1);
//
//     MEOW_DUMP_STATE("PostMix", xmm0, xmm1, xmm2, xmm3, xmm4, xmm5, xmm6, xmm7, 0);
//
//     paddq(xmm0, xmm2);
//     paddq(xmm1, xmm3);
//     paddq(xmm4, xmm6);
//     paddq(xmm5, xmm7);
//     pxor(xmm0, xmm1);
//     pxor(xmm4, xmm5);
//     paddq(xmm0, xmm4);
//
//     MEOW_DUMP_STATE("PostFold", xmm0, xmm1, xmm2, xmm3, xmm4, xmm5, xmm6, xmm7, 0);
//
//     return(xmm0);
// }
//
// // Streaming constructions removed
//
// #undef INSTRUCTION_REORDER_BARRIER
// #undef prefetcht0
// #undef movdqu
// #undef movdqu_mem
// #undef movq
// #undef aesdec
// #undef pshufb
// #undef pxor
// #undef paddq
// #undef pand
// #undef palignr
// #undef pxor_clear
// #undef MEOW_MIX
// #undef MEOW_MIX_REG
// #undef MEOW_SHUFFLE
// #undef MEOW_DUMP_STATE
//
// // MeowExpandSeed() removed
//
// #define MEOW_HASH_X64_AESNI_H
// #endif
import "C"
import (
	"encoding/binary"
	"fmt"
	"hash"
	"unsafe"
)

// Size constants.
const (
	SeedSize  = 128 // 128 bytes for a seed.
	HashSize  = 16  // 16 bytes for a hash.
	BlockSize = 256 // 256 byte blocks.

	// Version info.
	Version     = 5
	VersionName = "0.5/calico"
)

// MeowDefaultSeed is a seed to use.
//
// The default seed is now a "nothing-up-our-sleeves" number for good measure.
// You may verify that it is just an encoding of Pi.
var MeowDefaultSeed = [SeedSize]byte{
	0x32, 0x43, 0xF6, 0xA8, 0x88, 0x5A, 0x30, 0x8D,
	0x31, 0x31, 0x98, 0xA2, 0xE0, 0x37, 0x07, 0x34,
	0x4A, 0x40, 0x93, 0x82, 0x22, 0x99, 0xF3, 0x1D,
	0x00, 0x82, 0xEF, 0xA9, 0x8E, 0xC4, 0xE6, 0xC8,
	0x94, 0x52, 0x82, 0x1E, 0x63, 0x8D, 0x01, 0x37,
	0x7B, 0xE5, 0x46, 0x6C, 0xF3, 0x4E, 0x90, 0xC6,
	0xCC, 0x0A, 0xC2, 0x9B, 0x7C, 0x97, 0xC5, 0x0D,
	0xD3, 0xF8, 0x4D, 0x5B, 0x5B, 0x54, 0x70, 0x91,
	0x79, 0x21, 0x6D, 0x5D, 0x98, 0x97, 0x9F, 0xB1,
	0xBD, 0x13, 0x10, 0xBA, 0x69, 0x8D, 0xFB, 0x5A,
	0xC2, 0xFF, 0xD7, 0x2D, 0xBD, 0x01, 0xAD, 0xFB,
	0x7B, 0x8E, 0x1A, 0xFE, 0xD6, 0xA2, 0x67, 0xE9,
	0x6B, 0xA7, 0xC9, 0x04, 0x5F, 0x12, 0xC7, 0xF9,
	0x92, 0x4A, 0x19, 0x94, 0x7B, 0x39, 0x16, 0xCF,
	0x70, 0x80, 0x1F, 0x2E, 0x28, 0x58, 0xEF, 0xC1,
	0x66, 0x36, 0x92, 0x0D, 0x87, 0x15, 0x74, 0xE6,
}

// Hash data to 16 byte hash using MeowDefaultSeed.
func Hash(data []byte) []byte {
	m128i := C.MeowHash(
		unsafe.Pointer(&MeowDefaultSeed),
		C.ulonglong(len(data)),
		unsafe.Pointer(&data[0]))
	b := *(*[16]byte)(unsafe.Pointer(&m128i))
	return b[:]
}

// HashSeed hashes data using seed.
func HashSeed(seed [SeedSize]byte, data []byte) []byte {
	m128i := C.MeowHash(
		unsafe.Pointer(&seed),
		C.ulonglong(len(data)),
		unsafe.Pointer(&data[0]))
	b := *(*[16]byte)(unsafe.Pointer(&m128i))
	return b[:]
}

// String prints 4 32-bit chunks in hex. High bytes are on left.
// Panics if len(hash) is less than HashSize.
//
// The `%x` verb in package fmt prints identical hex, but with high
// bytes on the right.
func String(hash []byte) string {
	return fmt.Sprintf("%X-%X-%X-%X",
		binary.LittleEndian.Uint32(hash[12:16]),
		binary.LittleEndian.Uint32(hash[8:12]),
		binary.LittleEndian.Uint32(hash[4:8]),
		binary.LittleEndian.Uint32(hash[0:4]))
}

// meowHash implements hash.Hash using the Meow Hash function.
type meowHash []byte

// New makes a new hash.Hash using the meowHash type.
func New() hash.Hash { return &meowHash{} }

// Write p to h.
func (h *meowHash) Write(p []byte) (n int, err error) {
	*h = append(*h, p...)
	return len(p), nil
}

// Sum calls Hash() on the data written via Write(). The argument b is ignored.
// The returned slice is HashSize (16) bytes long.
func (h *meowHash) Sum(b []byte) []byte {
	x := Hash(*h)
	return x
}

// Reset erases the data accumulated via Write().
func (h *meowHash) Reset() { *h = (*h)[:0] }

// Size of hash.
func (h *meowHash) Size() int { return HashSize }

// BlockSize is the ideal size of blocks.
func (h *meowHash) BlockSize() int { return BlockSize }
