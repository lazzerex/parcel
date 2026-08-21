variable "local_emulator" {
  description = "Relax provider checks that a local AWS emulator cannot satisfy. Must stay false against real AWS."
  type        = bool
  default     = false
}
