.3ds
.thumb

.open "code.bin", "build/patched_code.bin", 0x100000

;;;
; Globals
;;;

snprintf equ 0x124b24

;;;
; SpotPass Patches
;
; @author RSM (https://github.com/giroletm / https://gitlab.com/giroletm)
; @research 3dbrew contributors (https://www.3dbrew.org) ; DaniElectra (https://github.com/DaniElectra) ; RSM (https://github.com/giroletm / https://gitlab.com/giroletm)
;;;

; The function at 0x0010B1BC ("BOSS_ConvertAakamaitoNPDL") patches the requested URL to fix legacy akamai links. Let's add our own URL patch to it!
; We save the first subdomain of the URL. Then, we check if it starts with "np".
; If so, then make sure the rest of the domain name matches with initialSpotpassUrl (cdn.nintendowifi.net).
; Finally, build a new URL with newSpotpassUrl (https://api.netpass.cafe/) followed by the saved subdomain and then the original path
; This way, we redirect SpotPass requests to our server.
; For example, "https://npdl.cdn.nintendowifi.net/some/path" should be turned into "https://api.netpass.cafe/npdl/some/path"

; For Nintendo Video support we need to turn "http://pubXX-p.est.c.app.nintendowifi.net/some/path" into "https://api.netpass.cafe/pubXX-p/some/path"

strncmp equ 0x125148
strncpy equ 0x126C08
strlen equ 0x12775c

URLBufferSizePtr equ 0x10B208
sub_10C104 equ 0x10C104
BOSS_MakeHTTPRequest equ 0x122A2C

; BOSS_ConvertAakamaitoNPDL has two endings, patch both to redirect to somewhere we have more space to write custom code

BOSS_ConvertAakamaitoNPDL_FirstEnding equ 0x10B1F2
BOSS_ConvertAakamaitoNPDL_SecondEnding equ 0x10B1FE

.org BOSS_ConvertAakamaitoNPDL_FirstEnding
.area 0x4
  bl ConvertAakamaitoNPDL_NewFirstEnding
.endarea

.org BOSS_ConvertAakamaitoNPDL_SecondEnding
.area 0x4
  bl ConvertAakamaitoNPDL_NewSecondEnding
.endarea

; We need to push call-safe registers to maintain their call safety.
; Then, just call whichever function each ending was originally calling.
; After that, we can push non-call-safe registers so we can restore them later to make sure the function we patched returns the same values
; Finally, call our URL patching function!

.org 0x11E0E4
.area 0x18
.db 0, 0 ; zero-termination of "string"
ConvertAakamaitoNPDL_NewFirstEnding:
  push {r0-r8, lr}
  bl sub_10C104 ; Call the original method
  bl UseCustomNex
  cmp r0, 0
  bne ConvertAakamaitoNPDL_NewFirstEnding_do
  pop {r0-r8, pc}
ConvertAakamaitoNPDL_NewFirstEnding_do:
  bl PatchSpotpassUrl
.endarea

.org 0x1225D0
.area 0x18
.db 0, 0 ; zero-termination of "string"
ConvertAakamaitoNPDL_NewSecondEnding:
  push {r0-r8, lr}
  blx strncpy
  bl UseCustomNex
  cmp r0, 0
  bne ConvertAakamaitoNPDL_NewSecondEnding_do
  pop {r0-r8, pc}
ConvertAakamaitoNPDL_NewSecondEnding_do:
  bl PatchSpotpassUrl
.endarea

; This function essentially:
; - Finds the index to the end of the protocol part of the URL (after "https://" in the case of an HTTPS request, for example)
;   - If there isn't any protocol, then consider it's automatic and just use index 0
; - From this, checks if the subdomain starts with "np"
;   - If not:
;     - Leave
;   - If yes:
;     - Find the index of the end of the first subdomain part of the URL.
;     - From this, save the subdomain into a buffer, and check if the rest of the domain is Nintendo's CDN
;       - If not:
;         - Leave
;       - If yes:
;         - Buffer the URL into a temporary buffer
;         - Build a new URL from our URL, the buffered subdomain, and the aforementionned temporary buffer

.org 0x10BBCC
.area 0x1C
.db 0, 0 ; zero-termination of "string"
PatchSpotpassUrl:
  ldr r6, [URLBufferSizePtrPtr]
  ldr r6, [r6]

  mov r1, r6 ; total stack size in r1
  add r1, 0x18 ; two snprintf parameters and subdomain buffer is 0x10 large

  ; We want 0x14 + (URL buffer size) bytes of temporary memory in the stack.
  ; Unfortunately, "sub sp, r6" is unsupported by the 3DS. As a solution, add the negative equivalent of r6
  neg r0, r1
  add sp, r0

  ; clear the stack
  ; r1 already holds the stack size
  add r0, sp, 0
  blx memclr
  ; We want to keep the pointer to the URL buffer and to the temporary buffer somewhere persistent

  ; r4 - protocol offset backup
  ; r5 - current replace object pointer
  ; r6 - input buffer size
  ; r7 - url buffer

  bl PatchSpotpassUrl_cont1

.align
URLBufferSizePtrPtr:
  .word URLBufferSizePtr ; The URL buffer's size is fixed and stored in memory. Let's get it from there instead of re-hardcoding it
.endarea

.org 0x12981C
.area 0x30
.db 0, 0 ; zero-termination of "string"
PatchSpotpassUrl_cont1:
  mov r4, #0 ; we put the protocol offset into r4
  ; Find the offset to the end of the protocol par of the URL

  ldrb r1, [r7, r4] ; Read the r4'th character of the URL buffer
  cmp r1, 0 ; Is it null?
  beq PatchSpotpassUrl_FindEndProtocolLoop_ExitNoProtocol

  b PatchSpotpassUrl_FindEndProtocolLoop_DoLoop ; Otherwise, start the loop

PatchSpotpassUrl_FindEndProtocolLoop_NextIteration:
  add r4, r4, 1 ; Increment the reusable index
  cmp r4, r6 ; Compare with the URL buffer size
  bge PatchSpotpassUrl_SkipPatch_Redirect1 ; If index >= URL buffer size, then we must break from the loop

PatchSpotpassUrl_FindEndProtocolLoop_DoLoop:
  add r1, r7, r4
  ldrb r1, [r1, 1] ; Read the (r4+1)'th character of the URL buffer

  cmp r1, '/' ; Is it a slash?
  beq PatchSpotpassUrl_FindEndProtocolLoop_NextIsSlash ; If so, go check for the rest

  cmp r1, 0 ; Is it null?
  beq PatchSpotpassUrl_FindEndProtocolLoop_ExitNoProtocol ; If so, there's no protocol in the URL, move on

  b PatchSpotpassUrl_FindEndProtocolLoop_NextIteration ; If neither or slash nor null, go to the next iteration

PatchSpotpassUrl_FindEndProtocolLoop_NextIsSlash:
  ldrb r1, [r7, r4] ; Read the r4'th character of the URL buffer
  cmp r1, '/' ; Is it a slash?
  bne PatchSpotpassUrl_FindEndProtocolLoop_NextIteration ; If not, move to the next iteration

  add r1, r4, 2 ; We just read two slashes in a row, so remember the index of the character right after it

PatchSpotpassUrl_FindEndProtocolLoop_ExitNoProtocol:
PatchSpotpassUrl_FindEndProtocolLoop_ExitEndLoop:
  bl PatchSpotpassUrl_cont2

PatchSpotpassUrl_SkipPatch_Redirect1:
  bl PatchSpotpassUrl_SkipPatch

.endarea

.org 0x12F0A0
.area 0x1C
.db 0, 0 ; zero-termination of "string"
PatchSpotpassUrl_cont2:
  mov r4, r1 ; Here, either r1 is 0 from reading a null character, or it is the index to the subdomain

  ; Now, r4 contains the offset to the subdomain part

  ldr r5, [spotpassUrlRewritePtr]
  
  b PatchSpotpassUrl_LoopEntry
PatchSpotpassUrl_LoopContinue:
  add r5, r5, 7 ; iterate to the next object
  add r5, r5, 5

PatchSpotpassUrl_LoopEntry:
  ; first, we check if the subdomain part matches
  ldr r0, [r5, 0x0] ; the subdomain start match
  cmp r0, 0

  bl PatchSpotpassUrl_cont3
.align
spotpassUrlRewritePtr:
  .word spotpassUrlRewrite
.endarea

.org 0x113694
.area 0x24
.db 0, 0 ; zero-termination of "string"
PatchSpotpassUrl_cont3:
  beq PatchSpotpassUrl_SkipPatch_Redirect2 ; if we already checked all then do nothing further
  blx strlen
  
  mov r8, r0 ; back up the length of the subdomain match
  mov r2, r0
  add r0, r7, r4 ; get the start of the url
  ldr r1, [r5, 0x0] ; subdomain prefix to match against
  blx strncmp
  cmp r0, 0
  bne PatchSpotpassUrl_LoopContinue_Redirect1 ; no match

  bl PatchSpotpassUrl_cont4
PatchSpotpassUrl_SkipPatch_Redirect2:
  bl PatchSpotpassUrl_SkipPatch
PatchSpotpassUrl_LoopContinue_Redirect1:
  bl PatchSpotpassUrl_LoopContinue
.endarea

.org 0x113818
.area 0x2C
.db 0, 0 ; zero-termination of "string"
PatchSpotpassUrl_cont4:
  ; ok, we now know that the subdomain part matches. so, let's loop
  ; to the end of the subdomain
  mov r0, r8 ; length of subdomain match
  add r0, r0, r4 ; r0 now has our current counter / offset
PatchSpotpassUrl_FindEndSubdomainLoop_Continue:
  add r0, r0, 1
  ; if the next char is 0 or / we have nothing to patch
  ldrb r1, [r7, r0]
  cmp r1, 0
  beq PatchSpotpassUrl_LoopContinue_Redirect2
  cmp r1, '/'
  beq PatchSpotpassUrl_LoopContinue_Redirect2
  cmp r1, '.'
  bne PatchSpotpassUrl_FindEndSubdomainLoop_Continue
  add r0, r0, 1 ; skip the dot

  ; now we copy the subdomain buffer onto the backup
  mov r8, r0 ; back up the index to the end of the sub domain
  add r1, r7, r4 ; url buffer + subdomain index
  sub r2, r0, r4 ; end of subdomain index - subdomain index
  add r0, sp, 8 ; subdomain buffer
  bl PatchSpotpassUrl_cont5
PatchSpotpassUrl_LoopContinue_Redirect2:
  bl PatchSpotpassUrl_LoopContinue
.endarea

.org 0x1138D8
.area 0x34
.db 0, 0 ; zero-termination of "string"
PatchSpotpassUrl_cont5:
  blx strncpy
  ; now it is time to check if the domain part matches up
  ldr r0, [r5, 0x4]
  blx strlen


  mov r2, r0
  ldr r1, [r5, 0x4]
  mov r3, r8 ; offset to start of subdomain
  add r0, r3, r7
  
  add r3, r3, r2
  mov r8, r3 ; save the full offset into r8
  
  blx strncmp
  cmp r0, 0
  bne PatchSpotpassUrl_LoopContinue_Redirect2
  ; ok, it does start with the correct domain as well, so we have a url to patch!

  ; Copy the path to the temporary buffer
  mov r3, r8
  add r0, sp, 0x18 ; temporary buffer
  add r1, r7, r3 ; path of the url
  mov r2, r6 ; size of the url / temporary buffer
  blx strncpy

  bl PatchSpotpassUrl_cont6
.endarea

.org 0x111878
.area 0x2C
.db 0, 0 ; zero-termination of "string"
PatchSpotpassUrl_cont6:
  ; now time to merge all the URL parts together
  mov r0, r7 ; the url buffer
  mov r1, r6 ; buffer size
  ldr r2, [newSpotpassUrlPatternPtr] ; pattern
  ldr r3, [r5, 0x8] ; path prefix
  add r4, sp, 8 ; subdomain buffer
  str r4, [sp, 0]
  add r4, sp, 0x18 ; path buffer
  str r4, [sp, 4]
  bl snprintf
PatchSpotpassUrl_SkipPatch:
  ; Move the stack pointer to its original position
  
  add sp, r6
  add sp, 0x18
  
  ; Restore the registers we've saved and return to the patched function
  
  pop {r0-r8, pc}
.align
newSpotpassUrlPatternPtr:
  .word newSpotpassUrlPattern
.endarea

.org 0x148a7c
.area 0x200

spotpassUrlRewrite:
  .word spotpassGeneralPrefix
  .word spotpassGeneralUrl
  .word spotpassGeneralPath
  .word 0
  .word spotpassVideoPrefix
  .word spotpassVideoUrl
  .word spotpassVideoPath
  
  .word 0 

spotpassGeneralPrefix:
  .asciiz "np"
spotpassGeneralUrl:
  .asciiz "c.app.nintendowifi.net"
spotpassGeneralPath:
  .asciiz "/"

spotpassVideoPrefix:
  .asciiz "pub"
spotpassVideoUrl:
  .asciiz "est.c.app.nintendowifi.net"
spotpassVideoPath:
  .asciiz "/"

.align 4
newSpotpassUrlPattern:
  .asciiz "https://10.0.0.49%s%s%s"
.endarea