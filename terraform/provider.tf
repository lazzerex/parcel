terraform {
  required_version = ">= 1.10"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

# Region, credentials, and endpoint all come from the standard AWS environment
# chain, so no secret is committed and nothing here names the emulator.
# The relaxed checks below default to off and are enabled only by a local,
# untracked tfvars file.
provider "aws" {
  skip_credentials_validation = var.local_emulator
  skip_metadata_api_check     = var.local_emulator
  skip_requesting_account_id  = var.local_emulator
  s3_use_path_style           = var.local_emulator
}
