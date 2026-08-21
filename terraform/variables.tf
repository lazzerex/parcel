variable "local_emulator" {
  description = "Relax provider checks that a local AWS emulator cannot satisfy. Must stay false against real AWS."
  type        = bool
  default     = false
}

variable "project_name" {
  description = "Name prefix for every Parcel resource."
  type        = string
  default     = "parcel"
}

locals {
  bucket_name = "${var.project_name}-files"
  table_name  = "${var.project_name}-metadata"
  queue_name  = "${var.project_name}-jobs"
  dlq_name    = "${var.project_name}-jobs-dlq"
}
