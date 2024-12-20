variable "region" {
  description = "The AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

variable "name" {
  description = "The name of the S3 bucket"
  type        = string
  default     = null
}

variable "versioning" {
  description = "Enable versioning on the S3 bucket"
  type        = bool
  default     = false
}
